package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blockloop/scan/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v3"

	"github.com/wisdom-oss/service-consumers/globals"
	"github.com/wisdom-oss/service-consumers/types"
)

func CreateNewConsumer(w http.ResponseWriter, r *http.Request) {
	// get the error handler and the error handler status channel
	errorHandler := r.Context().Value(wisdomMiddleware.ERROR_CHANNEL_NAME).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.STATUS_CHANNEL_NAME).(<-chan bool)

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

	res, err := globals.SqlQueries.Query(tx, "insert-consumer",
		consumer.Name,
		consumer.Description,
		consumer.Address,
		consumer.Location,
		consumer.UsageType,
		consumer.AdditionalProperties,
	)
	if err != nil {
		log.Error().Err(err).Msg("unable to insert the consumer into the database")
		errorHandler <- fmt.Errorf("unable to insert the consumer into the database: %w", err)
		<-statusChannel
		tx.Rollback()
		return
	}

	// now get the id of the consumer
	var consumerID uuid.UUID
	err = scan.Row(&consumerID, res)
	if err != nil {
		log.Error().Err(err).Msg("unable to get the inserted consumer id")
		errorHandler <- fmt.Errorf("unable to get the inserted consumer id: %w", err)
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
	w.Header().Set("Location", fmt.Sprintf("./%s", consumerID.String()))
	w.WriteHeader(http.StatusCreated)
	err = tx.Commit()
}
