package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v3"

	"github.com/wisdom-oss/service-consumers/globals"
)

// SingleConsumer allows pulling one consumer with all their data attached
func SingleConsumer(w http.ResponseWriter, r *http.Request) {
	// get the error handler and the error handler status channel
	errorHandler := r.Context().Value(wisdomMiddleware.ERROR_CHANNEL_NAME).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.STATUS_CHANNEL_NAME).(<-chan bool)

	// now get the consumer id from the url
	consumerID := chi.URLParam(r, "consumer")

	// now build the sql query
	baseQuery, err := globals.SqlQueries.Raw("get-consumers")
	if err != nil {
		errorHandler <- fmt.Errorf("unable to build query: %w", err)
		<-statusChannel
		return
	}
	idFilter, err := globals.SqlQueries.Raw("filter-ids")
	if err != nil {
		errorHandler <- fmt.Errorf("unable to build query: %w", err)
		<-statusChannel
		return
	}

	// now merge the two filters
	sql := fmt.Sprintf(`%s WHERE %s`, strings.Trim(baseQuery, ";"), idFilter)

	// now prepare the query
	query, err := globals.Db.Prepare(sql)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to preparse database query: %w", err)
		<-statusChannel
		return
	}

	// now execute the sql query
	query.Query(pq.Array([]string{consumerID}))

}
