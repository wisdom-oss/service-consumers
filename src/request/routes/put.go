package routes

import (
	"encoding/json"
	"errors"
	"github.com/blockloop/scan/v2"
	"github.com/lib/pq"
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

	var usageType *string
	if newConsumerData.UsageType != nil {
		// try to determine the uuid of the usage type
		usageTypeRow, err := globals.Queries.QueryRow(connections.DbConnection, "get-consumer-type-id", newConsumerData.UsageType)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the usage type")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
		err = usageTypeRow.Scan(&usageType)
		if err != nil {
			l.Error().Err(err).Msg("failed to execute the update query for the usage type")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
	} else {
		usageType = nil
	}

	var jsonString *string
	if newConsumerData.AdditionalProperties != nil {
		jsonBytes, err := json.Marshal(newConsumerData.AdditionalProperties)
		if err != nil {
			l.Error().Err(err).Msg("failed to marshal the additional properties")
			e, _ := requestErrors.WrapInternalError(err)
			requestErrors.SendError(e, w)
			return
		}
		s := string(jsonBytes)
		jsonString = &s
	} else {
		jsonString = nil
	}

	// now execute the insertion query
	consumerIDRow, err := globals.Queries.Query(connections.DbConnection, "insert-consumer",
		newConsumerData.Name, newConsumerData.Coordinates[0], newConsumerData.Coordinates[1],
		usageType, jsonString)
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
	consumerRow, err := globals.Queries.Query(connections.DbConnection, "get-consumer-by-id", consumerID)
	if err != nil {
		l.Error().Err(err).Msg("failed to execute the query to get the consumer")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}
	// now access the query results
	var dbConsumer structs.DbConsumer
	err = scan.Row(&dbConsumer, consumerRow)

	if err != nil {
		l.Error().Err(err).Msg("failed to parse the database returns")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusCreated)

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
