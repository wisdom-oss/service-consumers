package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"microservice/errors"
	"microservice/helpers"
	"microservice/vars"
)

func AuthorizationCheck(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.WithFields(log.Fields{
			"middleware": true,
			"title":      "AuthorizationCheck",
		})
		logger.Debug("Checking the incoming request for authorization information set by the gateway")

		// Get the scopes the requesting user has
		scopes := r.Header.Get("X-Authenticated-Scope")
		// Check if the string is empty
		if strings.TrimSpace(scopes) == "" {
			logger.Warning("Unauthorized request detected. The required header had no content or was not set")
			requestError := errors.NewRequestError(errors.UnauthorizedRequest)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}

		scopeList := strings.Split(scopes, ",")
		if !helpers.StringArrayContains(scopeList, vars.Scope.ScopeValue) {
			logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
			requestError := errors.NewRequestError(errors.MissingScope)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}
		// Call the next handler which will continue handling the request
		nextHandler.ServeHTTP(w, r)
	})
}

/*
PingHandler

This handler is used to test if the service is able to ping itself. This is done to run a healthcheck on the container
*/
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

/*
RequestHandler

TODO: Write your own handler logic into this handler or use this handler as example
*/
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"title":      "RequestHandler",
	})
	// Check the method of the http request
	switch r.Method {
	case http.MethodGet:
		returnConsumerInformation(w, r)
		break
	case http.MethodPost:
		createNewConsumer(w, r)
		break
	case http.MethodPatch:
		updateConsumerInformation(w, r)
		break
	case http.MethodDelete:
		deleteConsumerFromDatabase(w, r)
		break
	default:
		requestError := errors.NewRequestError(errors.UnsupportedHTTPMethod)
		w.Header().Set("Content-Type", "text/json")
		w.WriteHeader(requestError.HttpStatus)
		encodingError := json.NewEncoder(w).Encode(requestError)
		if encodingError != nil {
			logger.WithError(encodingError).Error("Unable to encode the request error into json")
			return
		}
		break
	}
}

func returnConsumerInformation(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}

func createNewConsumer(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}

func updateConsumerInformation(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}

func deleteConsumerFromDatabase(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement code logic
}
