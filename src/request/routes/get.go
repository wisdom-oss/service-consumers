package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/blockloop/scan/v2"
	"github.com/lib/pq"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"net/http"
	"strconv"
)

// l is an alias for the http logger
var l = globals.HttpLogger

// GetConsumers handles request made to the "/" endpoint using the GET method
func GetConsumers(w http.ResponseWriter, r *http.Request) {
	l.Info().Msg("new request for consumer retrieval")
	// now try to retrieve the parameters `usage_above`, `id` and `in` from
	// the query parameters
	var usageAboveSet, consumerIdSet, areaFilterSet bool
	var usageAbove int
	var consumerIDs []string
	var areaFilter []string
	if usageAboveSet = r.URL.Query().Has("usage_above"); usageAboveSet {
		l.Debug().Msg("found usage filter. parsing into int")
		queryParameter := r.URL.Query().Get("usage_above")
		var err error
		usageAbove, err = strconv.Atoi(queryParameter)
		// since the value could not be converted into an integer send a error response back
		if err != nil {
			e, err := requestErrors.GetRequestError("INVALID_USAGE_AMOUNT")
			if err != nil {
				e, _ = requestErrors.WrapInternalError(err)
			}
			requestErrors.SendError(e, w)
			return
		}
	}
	if consumerIdSet = r.URL.Query().Has("id"); consumerIdSet {
		consumerIDs = r.URL.Query()["id"]
	}
	if areaFilterSet = r.URL.Query().Has("in"); areaFilterSet {
		areaFilter = r.URL.Query()["in"]
	}

	// now create variables which store query errors and returned rows
	var consumerRows *sql.Rows
	var queryError error

	// now switch through the filters
	switch {
	case usageAboveSet && consumerIdSet && areaFilterSet:
		l.Info().Str("filters", "usage,consumerID,areaFilter").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-usage-id-area",
			usageAbove, pq.Array(consumerIDs), pq.Array(areaFilter))
		break
	case usageAboveSet && consumerIdSet && !areaFilterSet:
		l.Info().Str("filters", "usage,consumerID").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-usage-id",
			usageAbove, pq.Array(consumerIDs))
		break
	case usageAboveSet && !consumerIdSet && areaFilterSet:
		l.Info().Str("filters", "usage,areaFilter").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-usage-area",
			usageAbove, pq.Array(areaFilter))
		break
	case !usageAboveSet && consumerIdSet && areaFilterSet:
		l.Info().Str("filters", "consumerID,areaFilter").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-id-area",
			pq.Array(consumerIDs), pq.Array(areaFilter))
		break
	case usageAboveSet && !consumerIdSet && !areaFilterSet:
		l.Info().Str("filters", "usage").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-usage",
			usageAbove)
		break
	case !usageAboveSet && !consumerIdSet && areaFilterSet:
		l.Info().Str("filters", "areaFilter").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-area",
			pq.Array(areaFilter))
		break
	case !usageAboveSet && consumerIdSet && !areaFilterSet:
		l.Info().Str("filters", "consumerID").Msg("querying database")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection,
			"get-consumers-by-id",
			pq.Array(consumerIDs))
		break
	case !usageAboveSet && !consumerIdSet && !areaFilterSet:
		l.Warn().Str("filters", "none").Msg("querying database without filters")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection, "get-all-consumers")
		break
	default:
		l.Error().Msg("unknown case of filters encountered. returning database contents")
		consumerRows, queryError = globals.Queries.Query(
			connections.DbConnection, "get-all-consumers")
		break
	}

	// since at least one query was executed now check for an error
	if queryError != nil {
		l.Error().Err(queryError).Msg("error during database query")
		e, _ := requestErrors.WrapInternalError(queryError)
		requestErrors.SendError(e, w)
		return
	}

	// now defer the closure of the sql connection/rows
	defer func(consumerRows *sql.Rows) {
		err := consumerRows.Close()
		if err != nil {
			l.Error().Err(err).Msg("error while closing the database connection")
		}
	}(consumerRows)

	// now iterate through the consumer rows and write the results to an array
	var dbConsumers []structs.DbConsumer
	err := scan.Rows(&dbConsumers, consumerRows)

	// if the length is 0 there are no consumers matching the filters
	if len(dbConsumers) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var consumers []structs.Consumer
	for _, c := range dbConsumers {
		consumer, _ := c.ToConsumer()
		consumers = append(consumers, *consumer)
	}

	// now return the collected data
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(consumers)
	if err != nil {
		l.Error().Err(err).Msg("unable to send response")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}
}
