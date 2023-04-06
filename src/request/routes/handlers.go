// Package routes
// This package contains all route handlers for the microservice
package routes

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

/*
BasicHandler

This handler shows how a basic handler works and how to send back a message
*/
func BasicHandler(responseWriter http.ResponseWriter, request *http.Request) {
	_, err := responseWriter.Write([]byte("Hello World"))
	if err != nil {
		log.WithError(err).Error("Unable to send response to the client")
	}
}
