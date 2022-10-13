package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	e "microservice/errors"
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
			requestError := e.NewRequestError(e.UnauthorizedRequest)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}

		scopeList := strings.Split(scopes, ",")
		if !helpers.StringArrayContains(scopeList, vars.ScopeConfiguration.ScopeValue) {
			logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
			requestError := e.NewRequestError(e.MissingScope)
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
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
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
			helpers.SendRequestError(e.InvalidQueryParameter, w)
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
				helpers.SendRequestError(e.InvalidQueryParameter, w)
				return
			}
		}
	}
	// HINT: The area keys do not need to be checked since these are simple strings which limit the areas in which the
	// 	consumers are searched. If none of the area keys match the request will return nothing
}

/*
GetConsumers

Return a response with the consumers matching the filters set by the query parameters.
*/
func GetConsumers(w http.ResponseWriter, r *http.Request) {
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
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	var consumers []structs.Consumer
	// Close the current open connection to the database
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.WithError(err).Error("Unable to close the rows from the database")
			helpers.SendRequestError(e.DatabaseQueryError, w)
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
			helpers.SendRequestError(e.DatabaseQueryError, w)
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
CreateNewConsumer

Create a new consumer according to the request bodies content
*/
func CreateNewConsumer(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"title":      "createNewConsumer",
	})
	// Read the body of the request and parse it into an object
	var newConsumerData structs.IncomingConsumerData
	parsingError := json.NewDecoder(r.Body).Decode(&newConsumerData)
	if parsingError != nil {
		helpers.SendRequestError(e.UnprocessableEntity, w)
		return
	}
	// Build the insertion query
	insertionQuery := `INSERT INTO water_usage.consumers VALUES ($1, st_makepoint($2, $3), default) RETURNING id`
	row := vars.PostgresConnection.QueryRow(insertionQuery, newConsumerData.Name,
		newConsumerData.Latitude, newConsumerData.Longitude)
	err := row.Err()
	if err != nil {
		var dbError *pq.Error
		errors.As(err, &dbError)
		if dbError.Code.Name() == "unique_violation" {
			logger.WithError(dbError).Warning("Unique constraint violation detected while inserting the new consumer.")
			helpers.SendRequestError(e.UniqueConstraintViolation, w)
			return
		} else {
			logger.WithError(dbError).Error("An error occurred while inserting the new consumer into the database")
			helpers.SendRequestError(e.DatabaseQueryError, w)
			return
		}
	}
	var consumerId string
	err = row.Scan(&consumerId)
	if err != nil {
		logger.WithError(err).Error("An error occurred while getting the id of the new consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	// Build a query to retrieve the just inserted consumer
	selectQuery := `SELECT id, name, st_asgeojson(location) FROM water_usage.consumers WHERE id = $1`
	consumerRow, selectError := vars.PostgresConnection.Query(selectQuery, consumerId)
	if selectError != nil {
		logger.WithError(selectError).Error("An error occurred while selecting the newly inserted consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	defer func(consumerRow *sql.Rows) {
		err := consumerRow.Close()
		if err != nil {
			logger.WithError(err).Error("Unable to close the rows from the database")
			helpers.SendRequestError(e.DatabaseQueryError, w)
			return
		}
	}(consumerRow)
	var consumerName string
	var consumerLocation geojson.Geometry
	for consumerRow.Next() {
		scanError := consumerRow.Scan(&consumerId, &consumerName, &consumerLocation)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while iterating through the result rows of the query.")
			helpers.SendRequestError(e.DatabaseQueryError, w)
			return
		}
		break
	}
	w.Header().Set("Content-Type", "application/json")
	encodingError := json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerId,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if encodingError != nil {
		logger.WithError(encodingError).Error("An error occurred while returning the response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func updateConsumer(updateConsumerName bool, updateConsumerLocation bool, newName string, newLat float64,
	newLong float64, consumerId string, w http.ResponseWriter) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"action":     "UpdateConsumerInformation",
	})
	var queryError error
	switch {
	case updateConsumerName && updateConsumerLocation:
		updateQuery := `UPDATE water_usage.consumers SET name = $1, location=st_makepoint($2, $3) WHERE id = $4`
		_, queryError = vars.PostgresConnection.Query(updateQuery, newName, newLat,
			newLong, consumerId)
		break
	case updateConsumerName && !updateConsumerLocation:
		updateQuery := `UPDATE water_usage.consumers SET name = $1 WHERE id = $2`
		_, queryError = vars.PostgresConnection.Query(updateQuery, newName, consumerId)

		break
	case !updateConsumerName && updateConsumerLocation:
		updateQuery := `UPDATE water_usage.consumers SET location=st_makepoint($2, $3) WHERE id = $4`
		_, queryError = vars.PostgresConnection.Query(updateQuery, newLat, newLong, consumerId)
		break
	default:
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if queryError != nil {
		logger.WithError(queryError).Error("An error occurred while updating the consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
}

func UpdateConsumerInformation(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"action":     "UpdateConsumerInformation",
	})
	pathVars := mux.Vars(r)
	consumerIdParameter := pathVars["consumer_id"]
	var consumerId, name string
	// Check if the consumer exists or if there is no consumer with this id
	consumerCheckQuery := `SELECT id, name FROM water_usage.consumers WHERE id = $1`
	consumerCheckRow := vars.PostgresConnection.QueryRow(consumerCheckQuery, consumerIdParameter)
	err := consumerCheckRow.Scan(&consumerId, &name)
	if err != nil && err == sql.ErrNoRows {
		logger.Warning("Trying to update a consumer which is not present in the database")
		helpers.SendRequestError(e.NoSuchConsumer, w)
		return
	} else if err != nil && err != sql.ErrNoRows {
		logger.WithError(err).Error("An error occurred while trying to check the database for the consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	var newConsumerData structs.IncomingConsumerData
	parsingError := json.NewDecoder(r.Body).Decode(&newConsumerData)
	if parsingError != nil {
		logger.WithError(parsingError).Warning("Detected unreadable json in request")
		helpers.SendRequestError(e.UnprocessableEntity, w)
		return
	}
	// Check which attributes shall be updated
	updateName := r.URL.Query().Has("update") && helpers.StringArrayContains(r.URL.Query()["update"], "name")
	updateLocation := r.URL.Query().Has("update") && helpers.StringArrayContains(r.URL.Query()["update"], "coordinates")
	updateConsumer(updateName, updateLocation, newConsumerData.Name, newConsumerData.Latitude,
		newConsumerData.Longitude, consumerId, w)

	selectQuery := `SELECT id, name, st_asgeojson(location) FROM water_usage.consumers WHERE id = $1`
	consumerRow, selectError := vars.PostgresConnection.Query(selectQuery, consumerId)
	if selectError != nil {
		logger.WithError(selectError).Error("An error occurred while selecting the newly inserted consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	defer func(consumerRow *sql.Rows) {
		err := consumerRow.Close()
		if err != nil {
			logger.WithError(err).Error("Unable to close the rows from the database")
			helpers.SendRequestError(e.DatabaseQueryError, w)
			return
		}
	}(consumerRow)
	var consumerName string
	var consumerLocation geojson.Geometry
	for consumerRow.Next() {
		scanError := consumerRow.Scan(&consumerId, &consumerName, &consumerLocation)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while iterating through the result rows of the query.")
			helpers.SendRequestError(e.DatabaseQueryError, w)
			return
		}
		break
	}
	w.Header().Set("Content-Type", "application/json")
	encodingError := json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerId,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if encodingError != nil {
		logger.WithError(encodingError).Error("An error occurred while returning the response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DeleteConsumerFromDatabase(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"action":     "DeleteConsumerFromDatabase",
	})
	pathVars := mux.Vars(r)
	consumerId := pathVars["consumer_id"]
	deleteQuery := `DELETE FROM water_usage.consumers WHERE id=$1`
	_, queryError := vars.PostgresConnection.Query(deleteQuery, consumerId)
	if queryError != nil {
		logger.WithError(queryError).Error("An error occurred while deleting the consumer")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}
