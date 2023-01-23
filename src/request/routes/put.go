package routes

import (
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	e "microservice/request/error"
	"microservice/structs"
	"microservice/vars"
	"net/http"
)

func CreateConsumer(w http.ResponseWriter, r *http.Request) {
	// configure a logger for this function to allow matching errors to this function
	logger := log.WithFields(log.Fields{
		"apiFunction": "CreateConsumer",
	})
	logger.Info("received new request for consumer information")
	logger.Debug("parsing the request parameters from the request")

	/// create a new object which will store the request body sent by the client
	var requestBody structs.RequestBody

	// now parse the incoming request body into the object
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		logger.WithError(err).Error("unable to parse request body")
		e.RespondWithInternalError(err, w)
		return
	}

	// now prepare the query used to insert the new location into the database
	queryText := `INSERT INTO water_usage.consumers VALUES ($1, st_makepoint($2, $3), default) RETURNING id`
	insertStatement, err := vars.PostgresConnection.Prepare(queryText)
	if err != nil {
		logger.WithError(err).Error("unable to prepare sql query")
		e.RespondWithInternalError(err, w)
		return
	}

	// now execute the prepared query and catch the returned consumer id
	consumerIdRow := insertStatement.QueryRow(requestBody.Name, requestBody.Latitude, requestBody.Longitude)

	// check if an error was returned by the query
	err = consumerIdRow.Err()
	if err != nil {
		// parse the error into database specific errors
		var databaseError *pq.Error
		errors.As(err, &databaseError)

		// now check if the unique constraint was violated
		if databaseError.Code.Name() == "unique_violation" {
			logger.Warning("user tried to insert a duplicate consumer")
			// send a request error back to the client
			requestError, err := e.BuildRequestError(e.UniqueConstraintViolation)
			if err != nil {
				logger.WithError(err).Error("unable to create request error")
				e.RespondWithInternalError(err, w)
				return
			}
			e.RespondWithRequestError(requestError, w)
			return
		} else {
			// since the database returned an error which cannot be handled by changing user input return an internal
			// error
			logger.WithError(err).Error("an error occurred while inserting the new consumer")
			e.RespondWithInternalError(err, w)
			return
		}
	}

	// now check for the new consumer id
	var consumerID string
	err = consumerIdRow.Scan(&consumerID)
	if err != nil {
		logger.WithError(err).Error("an error occurred while reading the query results")
		e.RespondWithInternalError(err, w)
		return
	}

	// prepare a new query which shall receive the information about the just created consumer
	queryText = `SELECT name, ST_ASGeoJSON(location) FROM water_usage.consumers  WHERE id = $1`
	selectStatement, err := vars.PostgresConnection.Prepare(queryText)
	if err != nil {
		logger.WithError(err).Error("unable to prepare select query")
		e.RespondWithInternalError(err, w)
		return
	}

	// now execute the query
	consumerRow := selectStatement.QueryRow(consumerID)

	// now check if an error occurred while fetching the consumer data
	err = consumerRow.Err()
	if err != nil {
		logger.WithError(err).Error("an error occurred while getting the consumer data from the database")
		e.RespondWithInternalError(err, w)
		return
	}

	// now access the query results
	var consumerName string
	var consumerLocation geojson.Geometry
	err = consumerRow.Scan(&consumerName, &consumerLocation)
	if err != nil {
		logger.WithError(err).Error("unable to retrieve query results")
		e.RespondWithInternalError(err, w)
		return
	}

	// now send the collected data back to the user
	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerID,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if err != nil {
		logger.WithError(err).Error("unable to encode response")
		e.RespondWithInternalError(err, w)
		return
	}
}
