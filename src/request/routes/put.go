package routes

import (
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"net/http"
)

func CreateConsumer(w http.ResponseWriter, r *http.Request) {
	l.Info().Msg("new consumer creation requested")

	// parse the request body sent in the request
	var newConsumerData structs.IncomingConsumer
	err := json.NewDecoder(r.Body).Decode(&newConsumerData)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse the request body")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	// now execute the insertion query
	consumerIDRow, err := globals.Queries.QueryRow(connections.DbConnection, "insert-consumer",
		newConsumerData.Name, newConsumerData.Latitude, newConsumerData.Longitude)
	if err != nil {
		var pqError *pq.Error
		errors.As(err, &pqError)

		if pqError.Code.Name() == "unique_violation" {
			l.Warn().Msg("duplicate consumer detected. cancelling insertion")
			e, err := requestErrors.GetRequestError("DUPLICATE_CONSUMER")
			if err != nil {
				e, _ = requestErrors.WrapInternalError(err)
				requestErrors.SendError(e, w)
				return
			}
			requestErrors.SendError(e, w)
			return
		} else {
			l.Error().Err(err).Msg("failed to execute the insertion query")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	}

	// now get the consumer id from the row
	var consumerID string
	err = consumerIDRow.Scan(&consumerID)
	if err != nil {
		l.Error().Err(err).Msg("failed to scan the consumer id from the row")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	// now get the consumer from the database
	consumerRow, err := globals.Queries.QueryRow(connections.DbConnection, "get-consumer-by-id", consumerID)
	if err != nil {
		l.Error().Err(err).Msg("failed to execute the query to get the consumer")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}
	// now access the query results
	var consumerName string
	var consumerLocation geojson.Geometry
	err = consumerRow.Scan(&consumerName, &consumerLocation)
	if err != nil {
		l.Error().Err(err).Msg("failed to scan the consumer from the row")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerID,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if err != nil {
		l.Error().Err(err).Msg("failed to encode the consumer data")
	}
}
