package routes

import (
	"github.com/go-chi/chi/v5"
	requestErrors "microservice/request/error"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"net/http"
)

func DeleteConsumer(w http.ResponseWriter, r *http.Request) {
	l.Info().Msg("consumer deletion requested")
	consumerID := chi.URLParam(r, "consumerID")

	_, err := globals.Queries.Exec(connections.DbConnection, "delete-consumer", consumerID)
	if err != nil {
		l.Error().Err(err).Msg("failed to execute the delete query")
		e, _ := requestErrors.WrapInternalError(err)
		requestErrors.SendError(e, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
