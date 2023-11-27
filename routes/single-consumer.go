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

// SingleConsumer allows pulling one consumer with all their data attached
func SingleConsumer(w http.ResponseWriter, r *http.Request) {
	// get the error handler and the error handler status channel
	errorHandler := r.Context().Value(wisdomMiddleware.ERROR_CHANNEL_NAME).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.STATUS_CHANNEL_NAME).(<-chan bool)

	// now get the consumer id from the url
	consumerID := chi.URLParam(r, "consumer-id")

	// now build the sql query
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

	// since the consumer has been successfully scanned, return it to the
	// user
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error().Err(err).Msg("unable to return consumer")
		errorHandler <- fmt.Errorf("unable to return json response: %w", err)
		<-statusChannel
		return
	}
}
