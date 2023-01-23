package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	e "microservice/request/error"
	"microservice/structs"
	"microservice/vars"
	"net/http"
)

func UpdateConsumer(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"apiFunction": "UpdateConsumer",
	})
	logger.Info("received new consumer update request")
	logger.Debug("parsing the query parameters for the request")
	// parse the query parameters using gorilla/schema
	queryParameters := new(structs.UpdateConsumerQueryParameters)
	err := schema.NewDecoder().Decode(queryParameters, r.URL.Query())
	if err != nil {
		logger.WithError(err).Error("unable to parse the query parameters")
		e.RespondWithInternalError(err, w)
		return
	}

	// now parse the request body
	var requestBody structs.RequestBody
	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		logger.WithError(err).Error("unable to parse the request body")
		e.RespondWithInternalError(err, w)
		return
	}

	// now access the consumer id which is set as a path variable
	pathParameters := mux.Vars(r)
	consumerID := pathParameters["consumer_id"]

	// now prepare a query to check if the consumer is in the database
	queryText := `SELECT name, ST_ASGeoJSON(location) FROM water_usage.consumers WHERE id = $1`
	selectStatement, err := vars.PostgresConnection.Prepare(queryText)
	if err != nil {
		logger.WithError(err).Error("an error occurred while preparing the query for checking the consumer existence")
		e.RespondWithInternalError(err, w)
		return
	}

	// now query the database for the consumer
	consumerRow := selectStatement.QueryRow(consumerID)

	// check if an error was returned by the database
	err = consumerRow.Err()
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warning("user tried to update a non-existent consumer")
			response, err := e.BuildRequestError(e.NoConsumerFound)
			if err != nil {
				logger.WithError(err).Error("an error occurred while creating the error response")
				e.RespondWithInternalError(err, w)
				return
			}
			e.RespondWithRequestError(response, w)
			return
		} else {
			logger.WithError(err).Error("an error occurred while querying the database for the supplied consumer")
			e.RespondWithInternalError(err, w)
			return
		}
	}

	// close the connection to the database
	defer func(s *sql.Stmt) {
		err := s.Close()
		if err != nil {
			logger.WithError(err).Error("unable to close connection to the database")
			e.RespondWithInternalError(err, w)
			return
		}
	}(selectStatement)

	// since the query returned a row, the consumer exists
	// check if the name shall be updated
	if queryParameters.UpdateName {
		// prepare the sql query
		queryText := `UPDATE water_usage.consumers SET name = $1 WHERE id = $2`
		updateStatement, err := vars.PostgresConnection.Prepare(queryText)
		if err != nil {
			logger.WithError(err).Error("unable to prepare name update query")
			e.RespondWithInternalError(err, w)
			return
		}

		// now execute the prepared query
		_, err = updateStatement.Exec(requestBody.Name, consumerID)

		if err != nil {
			logger.WithError(err).Error("an error occurred while executing the name update query")
			e.RespondWithInternalError(err, w)
			return
		}
	}

	if queryParameters.UpdateLocation {
		// prepare the sql query
		queryText := `UPDATE water_usage.consumers SET location = st_makepoint($1, $2) WHERE id = $3`
		updateStatement, err := vars.PostgresConnection.Prepare(queryText)
		if err != nil {
			logger.WithError(err).Error("unable to prepare location update query")
			e.RespondWithInternalError(err, w)
			return
		}

		// now execute the prepared query
		_, err = updateStatement.Exec(requestBody.Latitude, requestBody.Longitude, consumerID)

		if err != nil {
			logger.WithError(err).Error("an error occurred while executing the location update query")
			e.RespondWithInternalError(err, w)
			return
		}
	}

	// if neither of the two properties shall be changed send a 304 Not Modified back to the client
	if !queryParameters.UpdateName && !queryParameters.UpdateLocation {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// now query the updated consumer from the database
	consumerRow = selectStatement.QueryRow(consumerID)

	// check if an error occurred when executing the query
	err = consumerRow.Err()
	if err != nil {
		logger.WithError(err).Error("an error occurred while executing the query for the updated consumer")
		e.RespondWithInternalError(err, w)
		return
	}

	// get the information from the row
	var consumerName string
	var consumerLocation geojson.Geometry

	err = consumerRow.Scan(&consumerName, &consumerLocation)
	if err != nil {
		logger.WithError(err).Error("an error occurred while parsing the query results for the updated consumer")
		e.RespondWithInternalError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerID,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if err != nil {
		logger.WithError(err).Error("unable to return the response due to an encoding error")
		e.RespondWithInternalError(err, w)
		return
	}
}
