package routes

import (
	"encoding/json"
	"github.com/blockloop/scan/v2"
	"github.com/go-chi/chi/v5"
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

	if newConsumerData.Coordinates != nil {
		// now execute the update query
		_, err := globals.Queries.Exec(connections.DbConnection, "update-consumer-location", newConsumerData.Coordinates[0],
			newConsumerData.Coordinates[1], consumerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the location")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	}

	if newConsumerData.UsageType != nil {
		// try to determine the uuid of the usage type
		usageTypeRow, err := globals.Queries.QueryRow(connections.DbConnection, "get-consumer-type-id", newConsumerData.UsageType)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the usage type")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
		var usageType string
		err = usageTypeRow.Scan(&usageType)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the usage type")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
		// now execute the update query
		_, err = globals.Queries.Exec(connections.DbConnection, "update-consumer-usage-type", usageType, consumerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the usage type")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	}

	if newConsumerData.AdditionalProperties != nil {
		jsonBytes, err := json.Marshal(newConsumerData.AdditionalProperties)
		if err != nil {
			l.Error().Err(err).Msg("failed to marshal the additional properties")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
		jsonString := string(jsonBytes)
		_, err = globals.Queries.Exec(connections.DbConnection, "update-consumer-additional-properties", jsonString, consumerID)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the query to update the additional properties")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	}

	// now get the consumer from the database
	consumerRow, err := globals.Queries.Query(connections.DbConnection, "get-consumer-by-id", consumerID)
	if err != nil {
		l.Error().Err(err).Msg("failed to execute the query to get the consumer")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}
	var dbConsumer structs.DbConsumer
	err = scan.Row(&dbConsumer, consumerRow)

	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusOK)
	consumer, err := dbConsumer.ToConsumer()
	if err != nil {
		l.Error().Err(err).Msg("failed to execute the query to get the consumer")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		l.Error().Err(err).Msg("failed to encode the consumer data")
	}
}
