package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	geojson "github.com/paulmach/go.geojson"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"net/http"
)

func UpdateConsumer(w http.ResponseWriter, r *http.Request) {
	l.Info().Msg("new consumer update requested")
	consumerID := chi.URLParam(r, "consumerID")

	// parse the request body sent in the request
	var newConsumerData structs.IncomingConsumer
	err := json.NewDecoder(r.Body).Decode(&newConsumerData)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse the request body")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	if newConsumerData.Name != nil {
		// now execute the update query
		_, err := globals.Queries.Exec(connections.DbConnection, "update-consumer-name", newConsumerData.Name, consumerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the name")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	}

	if newConsumerData.Latitude != nil && newConsumerData.Longitude != nil {
		// now execute the update query
		_, err := globals.Queries.Exec(connections.DbConnection, "update-consumer-location", newConsumerData.Latitude, newConsumerData.Longitude, consumerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the location")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
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
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(structs.Consumer{
		UUID:     consumerID,
		Name:     consumerName,
		Location: consumerLocation,
	})
	if err != nil {
		l.Error().Err(err).Msg("failed to encode the consumer data")
	}
}
