package routes

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	e "microservice/request/error"
	"microservice/vars"
	"net/http"
)

func DeleteConsumer(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"apiFunction": "DeleteConsumer",
	})
	logger.Info("received new request for consumer information")
	logger.Debug("parsing the request parameters from the request")

	pathParameters := mux.Vars(r)
	consumerID := pathParameters["consumer_id"]

	// build the sql query
	queryText := `DELETE FROM water_usage.consumers WHERE id = $1`
	deleteStatement, err := vars.PostgresConnection.Prepare(queryText)

	if err != nil {
		logger.WithError(err).Error("unable to prepare deletion query")
		// send an internal error back to the client
		e.RespondWithInternalError(err, w)
		return
	}

	// now execute the query
	_, err = deleteStatement.Exec(consumerID)
	if err != nil {
		logger.WithError(err).Error("an error occurred executing the query")
		e.RespondWithInternalError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
