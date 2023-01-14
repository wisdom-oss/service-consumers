// Package routes
// This package contains all route handlers for the microservice
package routes

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

/*
PingHandler

This handler is used to test if the service is able to ping itself. This is done to run a healthcheck on the container
*/
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

/*
BasicHandler

This handler shows how a basic handler works and how to send back a message
*/
func BasicHandler(responseWriter http.ResponseWriter, request *http.Request) {
	logger := log.WithFields(
		log.Fields{
			"middleware": true,
			"title":      "BasicHandler",
		},
	)
	logger.WithField("request", request).Info("Got new request")
	_, err := responseWriter.Write([]byte("Hello World"))
	if err != nil {
		log.WithError(err).Error("Unable to send response to the client")
	}
}
