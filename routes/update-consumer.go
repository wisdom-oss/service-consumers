package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	// now parse the request body into the new consumer
	var consumer types.Consumer
	err := json.NewDecoder(r.Body).Decode(&consumer)
	if err != nil {
		log.Error().Err(err).Msg("unable to decode request body into consumer")
		errorHandler <- fmt.Errorf("unable to decode request body into consumer: %w", err)
		<-statusChannel
		return
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
