package middleware

import (
	log "github.com/sirupsen/logrus"
	e "microservice/request/error"
	"microservice/utils"
	"microservice/vars"
	"net/http"
	"strings"
)

func AuthorizationCheck(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			logger := log.WithFields(
				log.Fields{
					"middleware": true,
					"title":      "AuthorizationCheck",
				},
			)
			logger.Debug("Checking the incoming request for authorization information set by the gateway")
			if request.URL.Path == "/ping" {
				nextHandler.ServeHTTP(responseWriter, request)
				return
			}
			// Get the scopes the requesting user has
			scopes := request.Header.Get("X-Authenticated-Scope")
			// Check if the string is empty
			if strings.TrimSpace(scopes) == "" {
				logger.Warning("Unauthorized request detected. The required header had no content or was not set")
				err, _ := e.BuildRequestError(e.MissingAuthorizationInformation)
				e.RespondWithRequestError(err, responseWriter)
				return
			}

			scopeList := strings.Split(scopes, ",")
			if !utils.ArrayContains(scopeList, vars.ScopeConfiguration.ScopeValue) {
				logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
				err, _ := e.BuildRequestError(e.InsufficientScope)
				e.RespondWithRequestError(err, responseWriter)
				return
			}
			// Call the next handler which will continue handling the request
			nextHandler.ServeHTTP(responseWriter, request)
		},
	)
}
