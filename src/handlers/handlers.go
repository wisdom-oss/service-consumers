package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	"microservice/errors"
	"microservice/helpers"
	"microservice/structs"
	"microservice/vars"
)

func AuthorizationCheck(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.WithFields(log.Fields{
			"middleware": true,
			"title":      "AuthorizationCheck",
		})
		logger.Debug("Checking the incoming request for authorization information set by the gateway")

		// Get the scopes the requesting user has
		scopes := r.Header.Get("X-Authenticated-Scope")
		// Check if the string is empty
		if strings.TrimSpace(scopes) == "" {
			logger.Warning("Unauthorized request detected. The required header had no content or was not set")
			requestError := errors.NewRequestError(errors.UnauthorizedRequest)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}

		scopeList := strings.Split(scopes, ",")
		if !helpers.StringArrayContains(scopeList, vars.Scope.ScopeValue) {
			logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
			requestError := errors.NewRequestError(errors.MissingScope)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}
		// Call the next handler which will continue handling the request
		nextHandler.ServeHTTP(w, r)
	})
}

/*
PingHandler

This handler is used to test if the service is able to ping itself. This is done to run a healthcheck on the container
*/
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

/*
RequestHandler

This handler is used to dispatch the requests to the handlers respective to the request method.

Supported request methods are: GET, PATCH, PUT, DELETE
*/
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	// Check the method of the http request
	switch r.Method {
	case http.MethodGet:
		returnConsumerInformation(w, r)
		break
	case http.MethodPost:
		createNewConsumer(w, r)
		break
	case http.MethodPatch:
		updateConsumerInformation(w, r)
		break
	case http.MethodDelete:
		deleteConsumerFromDatabase(w, r)
		break
	default:
		helpers.SendRequestError(errors.UnsupportedHTTPMethod, w)
		break
	}
}

/*
validateRequestParameters

This function validates the query parameters:
	- usage_above: The minimal usage a consumer needs to have recorded once to be returned
	- id: The id of a consumer by which the consumer shall be searched by
	- in: The key of a shape sent by the geodata service to limit the search area
*/
func validateRequestParameters(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": true,
		"title":      "AuthorizationCheck",
	})
	// Check if the usage above parameter was set correctly if used
	usageAboveSet := r.URL.Query().Has("usage_above")
	if usageAboveSet {
		rawUsageAboveValue := r.URL.Query().Get("usage_above")
		if _, err := strconv.Atoi(rawUsageAboveValue); err != nil {
			logger.Warning("Found invalid value for 'usage_above' in request. Rejecting the request")
			helpers.SendRequestError(errors.InvalidQueryParameter, w)
			return
		}
	}
	// Check if the consumer ids have been set correctly
	consumerIdsSet := r.URL.Query().Has("id")
	if consumerIdsSet {
		consumerIds := r.URL.Query()["id"]
		for _, consumerId := range consumerIds {
			validUUID, _ := regexp.MatchString("^[0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4"+
				"}\\b-[0-9a-fA-F]{12}$", consumerId)
			if !validUUID {
				logger.Warning("Found invalid value for 'id' in request. Rejecting the request")
				helpers.SendRequestError(errors.InvalidQueryParameter, w)
				return
			}
		}
	}
	// HINT: The area keys do not need to be checked since these are simple strings which limit the areas in which the
	// 	consumers are searched. If none of the area keys match the request will return nothing
}

/*
returnConsumerInformation

Return a response with the consumers matching the filters set by the query parameters.
*/
func returnConsumerInformation(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"title":      "ReturnConsumerInformation",
	})
	// Validate the parameters for the get request if they have been set
	validateRequestParameters(w, r)
	// Create an empty database query string in preparation for the queries
	var databaseQuery string
	// Check which parameters are available for the database query
	usageAboveAvailable := r.URL.Query().Has("usage_above")
	consumerIdsAvailable := r.URL.Query().Has("id")
	areaKeysAvailable := r.URL.Query().Has("in")
	// Prepare objects for the query results
	var rows *sql.Rows
	var queryError error
	// Use a switch function to determine which query needs to be executed and query the database afterwards
	switch {
	case usageAboveAvailable && consumerIdsAvailable && areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1) AND
						       id = any($2) AND
						       ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($3)), location)`
		// Get the needed parameters from the query
		usageAbove, _ := strconv.Atoi(r.URL.Query().Get("usage_above"))
		consumerIds := r.URL.Query()["id"]
		areaKeys := r.URL.Query()["in"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(
			databaseQuery, usageAbove, pq.Array(consumerIds), pq.Array(areaKeys),
		)
		break
	case usageAboveAvailable && consumerIdsAvailable && !areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1) AND
						       id = any($2)`
		// Get the needed parameters from the query
		usageAbove, _ := strconv.Atoi(r.URL.Query().Get("usage_above"))
		consumerIds := r.URL.Query()["id"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, usageAbove, pq.Array(consumerIds))
		break
	case usageAboveAvailable && !consumerIdsAvailable && areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1) AND
						       ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($2)), location)`
		// Get the needed parameters from the query
		usageAbove, _ := strconv.Atoi(r.URL.Query().Get("usage_above"))
		areaKeys := r.URL.Query()["in"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, usageAbove, pq.Array(areaKeys))
		break
	case !usageAboveAvailable && consumerIdsAvailable && areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id = any($1) AND
						       ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($2)), location)`
		// Get the needed parameters from the query
		consumerIds := r.URL.Query()["id"]
		areaKeys := r.URL.Query()["in"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, pq.Array(consumerIds), pq.Array(areaKeys))
		break
	case usageAboveAvailable && !consumerIdsAvailable && !areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1)`
		// Get the needed parameters from the query
		usageAbove, _ := strconv.Atoi(r.URL.Query().Get("usage_above"))
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, usageAbove)
		break
	case !usageAboveAvailable && consumerIdsAvailable && !areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE id = any($1)`
		// Get the needed parameters from the query
		consumerIds := r.URL.Query()["id"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, pq.Array(consumerIds))
		break
	case !usageAboveAvailable && !consumerIdsAvailable && areaKeysAvailable:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers
						 WHERE ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($1)), location)`
		// Get the needed parameters from the query
		areaKeys := r.URL.Query()["in"]
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery, pq.Array(areaKeys))
		break
	default:
		databaseQuery = `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers`
		logger.Info("Executing a query in the database")
		rows, queryError = vars.PostgresConnection.Query(databaseQuery)
		break
	}
	if queryError != nil {
		logger.WithError(queryError).Error("An error occurred during the database query")
		helpers.SendRequestError(errors.DatabaseQueryError, w)
		return
	}
	var consumers []structs.Consumer
	// Close the current open connection to the database
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.WithError(err).Error("Unable to close the rows from the database")
			helpers.SendRequestError(errors.DatabaseQueryError, w)
			return
		}
	}(rows)
	// Iterate through the rows and construct the consumer objects
	for rows.Next() {
		var uuid string
		var name string
		var location geojson.Geometry

		scanError := rows.Scan(&uuid, &name, &location)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while iterating through the result rows of the query.")
			helpers.SendRequestError(errors.DatabaseQueryError, w)
			return
		}

		consumers = append(consumers, structs.Consumer{
			UUID:     uuid,
			Name:     name,
			Location: location,
		})
	}
	// Now check the length of the consumers array
	if len(consumers) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Since the query has found a consumer set the header to json and return the response
	w.Header().Set("Content-Type", "application/json")
	encodingError := json.NewEncoder(w).Encode(consumers)
	if encodingError != nil {
		logger.WithError(encodingError).Error("An error occurred while returning the response")
		w.WriteHeader(http.StatusInternalServerError)
	}

}

/*
Create a new consumer according to the request bodies content
*/
func createNewConsumer(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}

func updateConsumerInformation(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}

func deleteConsumerFromDatabase(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}
