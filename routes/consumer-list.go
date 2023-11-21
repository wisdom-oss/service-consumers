package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/blockloop/scan/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v3"

	"github.com/wisdom-oss/service-consumers/globals"
	"github.com/wisdom-oss/service-consumers/types"
)

// ConsumerList provides the handling of requests that will return a
// list of consumers.
// The list can be filtered by using the following query parameters:
//   - in
//   - id
//   - usageAbove
func ConsumerList(w http.ResponseWriter, r *http.Request) {
	// get the error handler and the error handler status channel
	errorHandler := r.Context().Value(wisdomMiddleware.ERROR_CHANNEL_NAME).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.STATUS_CHANNEL_NAME).(<-chan bool)

	// get the different sql parameters
	shapeKeys, shapeKeysSet := r.URL.Query()["in"]
	consumerIDs, consumerIDsSet := r.URL.Query()["id"]
	minimalUsages, minimalUsagesSet := r.URL.Query()["usageAbove"]

	/*
			The following check is only done to issue a deprecation warning when
			using the API in a deprecated way.
			In this case, the deprecated function is selecting a single consumer
		    using the id
	*/
	if consumerIDsSet && len(consumerIDs) == 1 {
		w.Header().Set("Warning", `299 consumer-management "Selecting a single consumer using the id filter is deprecated. Please use the /{consumer-id} endpoint"`)
	}

	// now build the sql using the pulled sql parameters
	sql, err := globals.SqlQueries.Raw("get-consumers")
	if err != nil {
		log.Error().Err(err).Msg("unable to build sql")
		errorHandler <- fmt.Errorf("unable to build sql: %w", err)
		<-statusChannel
		return
	}
	activeFilters := 0
	var arguments []interface{}
	// now check every filter option if they have been specified
	if shapeKeysSet {
		activeFilters++
		filter, err := globals.SqlQueries.Raw("filter-location")
		if err != nil {
			log.Error().Err(err).Msg("unable to load filter sql")
			errorHandler <- fmt.Errorf("unable to load filter sql: %w", err)
			<-statusChannel
			return
		}
		filter = strings.ReplaceAll(filter, "$1", fmt.Sprintf("$%d", activeFilters))
		if !strings.Contains(sql, "WHERE") {
			sql = strings.Trim(sql, ";")
			sql += fmt.Sprintf(" WHERE %s", filter)
		} else {
			sql += fmt.Sprintf(" AND %s", filter)
		}
		arguments = append(arguments, pq.Array(shapeKeys))
	}

	if consumerIDsSet {
		activeFilters++
		for _, consumerID := range consumerIDs {
			_, err = uuid.Parse(consumerID)
			if err != nil {
				errorHandler <- "INVALID_UUID_IN_FILTER"
				<-statusChannel
				return
			}
		}
		filter, err := globals.SqlQueries.Raw("filter-ids")
		if err != nil {
			log.Error().Err(err).Msg("unable to load filter sql")
			errorHandler <- fmt.Errorf("unable to load filter sql: %w", err)
			<-statusChannel
			return
		}
		filter = strings.ReplaceAll(filter, "$1", fmt.Sprintf("$%d", activeFilters))
		if !strings.Contains(sql, "WHERE") {
			sql = strings.Trim(sql, ";")
			sql += fmt.Sprintf(" WHERE %s", filter)
		} else {
			sql += fmt.Sprintf(" AND %s", filter)
		}
		arguments = append(arguments, pq.Array(consumerIDs))
	}

	if minimalUsagesSet {
		activeFilters++
		minimalUsageString := minimalUsages[0]
		minimalUsage, err := strconv.ParseFloat(minimalUsageString, 64)
		if err != nil {
			errorHandler <- "USAGE_AMOUNT_NAN"
			<-statusChannel
			return
		}
		filter, err := globals.SqlQueries.Raw("filter-usage-amount")
		if err != nil {
			log.Error().Err(err).Msg("unable to load filter sql")
			errorHandler <- fmt.Errorf("unable to load filter sql: %w", err)
			<-statusChannel
			return
		}
		filter = strings.ReplaceAll(filter, "$1", fmt.Sprintf("$%d", activeFilters))
		if !strings.Contains(sql, "WHERE") {
			sql = strings.Trim(sql, ";")
			sql += fmt.Sprintf(" WHERE %s", filter)
		} else {
			sql += fmt.Sprintf(" AND %s", filter)
		}
		arguments = append(arguments, minimalUsage)
	}

	// now prepare the query
	sql = strings.ReplaceAll(sql, ";", "")
	sql += ";"
	query, err := globals.Db.Prepare(sql)
	if err != nil {
		log.Error().Err(err).Msg("unable to preparse database query")
		errorHandler <- fmt.Errorf("unable to prepare database query: %w", err)
		<-statusChannel
		return
	}

	rows, err := query.Query(arguments...)
	if err != nil {
		log.Error().Err(err).Msg("unable to query database")
		errorHandler <- fmt.Errorf("unable to query database: %w", err)
		<-statusChannel
		return
	}

	// now scan the results
	var consumers []types.Consumer
	err = scan.Rows(&consumers, rows)
	if err != nil {
		log.Error().Err(err).Msg("unable to scan query results")
		errorHandler <- fmt.Errorf("unable to scan query results: %w", err)
		<-statusChannel
		return
	}

	if len(consumers) == 0 {
		// since there are no consumers that match the filters, return
		// 204 No Content as response
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// now return the consumers
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(consumers)
	if err != nil {
		log.Error().Err(err).Msg("unable to encode consumers into json")
		errorHandler <- fmt.Errorf("unable to encode consumers into json: %w", err)
		<-statusChannel
		return
	}

}
