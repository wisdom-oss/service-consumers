package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/schema"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	e "microservice/request/error"
	"microservice/structs"
	"microservice/vars"
	"net/http"
)

// GetConsumers allows the user to query the microservice for consumers matching the available filters
func GetConsumers(w http.ResponseWriter, r *http.Request) {
	// configure a logger for this function to allow matching errors to this function
	logger := log.WithFields(log.Fields{
		"apiFunction": "GetConsumers",
	})
	logger.Info("received new request for consumer information")
	logger.Debug("parsing the request parameters from the request")
	params := new(structs.QueryParameters)
	err := schema.NewDecoder().Decode(params, r.URL.Query())
	// check if the decoding of the query parameters worked
	if err != nil {
		// send an internal error back to the client
		e.RespondWithInternalError(err, w)
	}
	// now check which parameters have been set by testing the array lengths of the param object properties
	usageAboveSet := len(params.UsageAbove) > 0
	consumerIDSet := len(params.ConsumerIds) > 0
	areaFilterSet := len(params.AreaKeys) > 0

	// now create some empty objects in which the query results and errors are stored if an error should occur
	var queryRows *sql.Rows
	var queryError error
	var statement *sql.Stmt
	// now switch through all cases to determine the actual query to be executed
	switch {
	// all filter options have been supplied
	case usageAboveSet && consumerIDSet && areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
              	      WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1)
                      AND id = any($2)
                      AND st_contains((SELECT geom FROM geodata.shapes WHERE key = any($3)), location)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.UsageAbove[0], params.ConsumerIds, params.AreaKeys)
		// end the handling of the results here since the handling will be centralized
		break
	// only the query parameter "usage_above" and "id" have been set in the query
	case usageAboveSet && consumerIDSet && !areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
              	      WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1)
                      AND id = any($2)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.UsageAbove[0], params.ConsumerIds)
		// end the handling of the results here since the handling will be centralized
		break
	// the query parameters "usage_above" and "in" have been set in the request
	case usageAboveSet && !consumerIDSet && areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
              	      WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1)
                      AND st_contains((SELECT geom FROM geodata.shapes WHERE key = any($2)), location)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.UsageAbove[0], params.AreaKeys)
		// end the handling of the results here since the handling will be centralized
		break
	// the query parameters "id" and "in" have been set in the request
	case !usageAboveSet && consumerIDSet && areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
                      WHERE id = any($1)
                      AND st_contains((SELECT geom FROM geodata.shapes WHERE key = any($2)), location)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.ConsumerIds, params.AreaKeys)
		// end the handling of the results here since the handling will be centralized
		break
	// only the query parameter "id" has been set in the request
	case !usageAboveSet && consumerIDSet && !areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
                      WHERE id = any($1)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.ConsumerIds)
		// end the handling of the results here since the handling will be centralized
		break
	// only the query parameter "in" set has been set
	case !usageAboveSet && !consumerIDSet && areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
                      WHERE st_contains((SELECT geom FROM geodata.shapes WHERE key = any($1)), location)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.AreaKeys)
		// end the handling of the results here since the handling will be centralized
		break
	// only the query parameter "usage_above" has been set
	case usageAboveSet && !consumerIDSet && !areaFilterSet:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers 
              	      WHERE id IN (SELECT consumer FROM water_usage.usages WHERE value > $1)`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query(params.UsageAbove[0])
		// end the handling of the results here since the handling will be centralized
		break
	// no parameters were set
	default:
		queryText := `SELECT id, name, ST_ASGeoJSON(location) FROM water_usage.consumers`
		logger.Debug("preparing consumer query")
		// prepare the query statement to protect it against query injection
		statement, err = vars.PostgresConnection.Prepare(queryText)
		logger.Info("executing prepared query")
		queryRows, queryError = statement.Query()
		// end the handling of the results here since the handling will be centralized
		break
	}
	// now check if an error occurred during the query execution
	if queryError != nil {
		logger.WithError(queryError).Error("a query error occurred")
		e.RespondWithInternalError(queryError, w)
		// finish the handler
		return
	}

	// close the database connection to ensure no hangups or connection limit in the database
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			logger.WithError(err).Error("an error occurred while closing the database connection")
			e.RespondWithInternalError(err, w)
			return
		}
	}(queryRows)

	// now create a new array which will contain the consumers
	var consumers []structs.Consumer

	// iterate through the rows that were returned by the query execution
	for queryRows.Next() {
		// create variables into which the row shall be parsed
		var uuid, name string
		var location geojson.Geometry

		// try scanning the result row for the variables
		err := queryRows.Scan(&uuid, &name, &location)
		if err != nil {
			logger.WithError(err).Error("an error occurred while getting results from the query result")
			e.RespondWithInternalError(err, w)
			return
		}

		// append the collected data to the consumers array
		consumers = append(consumers, structs.Consumer{
			UUID:     uuid,
			Name:     name,
			Location: location,
		})
	}
	// check the length of the created array
	if len(consumers) == 0 {
		// since no consumers were found the service will send a No Content header
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// return the array as json
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(consumers)
	if err != nil {
		logger.WithError(err).Error("unable to encode the response")
		e.RespondWithInternalError(err, w)
		return
	}
}
