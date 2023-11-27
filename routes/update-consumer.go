package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/blockloop/scan/v2"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v3"

	"github.com/wisdom-oss/service-consumers/globals"
	"github.com/wisdom-oss/service-consumers/types"
)

func UpdateConsumer(w http.ResponseWriter, r *http.Request) {
	// get the error handler and the error handler status channel
	errorHandler := r.Context().Value(wisdomMiddleware.ERROR_CHANNEL_NAME).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.STATUS_CHANNEL_NAME).(<-chan bool)

	// get the id of the consumer that shall be updated
	consumerID := chi.URLParam(r, "consumer-id")

	// now get the consumer that has the id
	baseQuery, err := globals.SqlQueries.Raw("get-consumers")
	if err != nil {
		log.Error().Err(err).Msg("unable to build query")
		errorHandler <- fmt.Errorf("unable to build query: %w", err)
		<-statusChannel
		return
	}
	idFilter, err := globals.SqlQueries.Raw("filter-ids")
	if err != nil {
		log.Error().Err(err).Msg("unable to build query")
		errorHandler <- fmt.Errorf("unable to build query: %w", err)
		<-statusChannel
		return
	}

	// now merge the two filters
	sql := fmt.Sprintf(`%s WHERE %s`, strings.Trim(baseQuery, ";"), idFilter)

	// now prepare the query
	query, err := globals.Db.Prepare(sql)
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare database query")
		errorHandler <- fmt.Errorf("unable to preparse database query: %w", err)
		<-statusChannel
		return
	}

	// now execute the sql query
	rows, err := query.Query(pq.Array([]string{consumerID}))
	if err != nil {
		log.Error().Err(err).Msg("unable to query the database")
		errorHandler <- fmt.Errorf("unable to query database: %w", err)
		<-statusChannel
		return
	}

	// now scan the query results into a single consumer
	var consumer types.Consumer
	err = scan.Row(&consumer, rows)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse database query results")
		errorHandler <- fmt.Errorf("unable to parse query result: %w", err)
		<-statusChannel
		return
	}

	// now parse the request body into the new consumer
	var updatedConsumerRepresentation types.Consumer
	err = json.NewDecoder(r.Body).Decode(&updatedConsumerRepresentation)
	if err != nil {
		log.Error().Err(err).Msg("unable to decode request body into consumer")
		errorHandler <- fmt.Errorf("unable to decode request body into consumer: %w", err)
		<-statusChannel
		return
	}

	// now check which fields need to be updated
	if consumer.Name != updatedConsumerRepresentation.Name {
		consumer.Name = updatedConsumerRepresentation.Name
	}
	if consumer.Description != updatedConsumerRepresentation.Description {
		consumer.Description = updatedConsumerRepresentation.Description
	}
	if consumer.Address != updatedConsumerRepresentation.Address {
		consumer.Address = updatedConsumerRepresentation.Address
	}
	if consumer.Location != updatedConsumerRepresentation.Location {
		consumer.Location = updatedConsumerRepresentation.Location
	}
	if consumer.UsageType != updatedConsumerRepresentation.UsageType {
		consumer.UsageType = updatedConsumerRepresentation.UsageType
	}
	if consumer.AdditionalProperties != updatedConsumerRepresentation.AdditionalProperties {
		consumer.AdditionalProperties = updatedConsumerRepresentation.AdditionalProperties
	}

	// now write the consumer into the database
	tx, err := globals.Db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Error().Err(err).Msg("unable to start database transaction")
		errorHandler <- fmt.Errorf("unable to start database transaction: %w", err)
		<-statusChannel
		return
	}

	_, err = globals.SqlQueries.Exec(tx, "update-consumer",
		consumer.Name,
		consumer.Description,
		consumer.Address,
		consumer.Location,
		consumer.UsageType,
		consumer.AdditionalProperties,
		consumerID,
	)
	if err != nil {
		log.Error().Err(err).Msg("unable to insert the consumer into the database")
		errorHandler <- fmt.Errorf("unable to insert the consumer into the database: %w", err)
		<-statusChannel
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("unable to commit changes to the database")
		errorHandler <- fmt.Errorf("unable to commit changes to the database: %w", err)
		<-statusChannel
		tx.Rollback()
		return
	}

	// now set the location header and indicate that the consumer has been
	// created
	w.WriteHeader(http.StatusOK)
	err = tx.Commit()
}
