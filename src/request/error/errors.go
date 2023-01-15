// Package requestErrors contains all request errors which are directly handled by the handlers and are detected by
// the handlers. The request errors are identified by a constant value which also represents the error code
package requestErrors

import (
	"net/http"
)

const MissingAuthorizationInformation = "MISSING_AUTHORIZATION_INFORMATION"
const InsufficientScope = "INSUFFICIENT_SCOPE"
const InternalError = "INTERNAL_ERROR"
const UniqueConstraintViolation = "UNIQUE_CONSTRAINT_VIOLATION"
const NoConsumerFound = "NO_CONSUMER_FOUND"

var titles = map[string]string{
	MissingAuthorizationInformation: "Unauthorized",
	InsufficientScope:               "Insufficient Scope",
	InternalError:                   "Internal Error",
	UniqueConstraintViolation:       "Unique Constraint Violation",
	NoConsumerFound:                 "No Consumer Found",
}

var descriptions = map[string]string{
	MissingAuthorizationInformation: "the accessed resource requires authorization, " +
		"however the request did not contain valid authorization information. please check the request",
	InsufficientScope: "hhe authorization was successful, " +
		"but the resource is protected by a scope which was not included in the authorization information",
	InternalError:             "during the handling of the request an unexpected error occurred",
	UniqueConstraintViolation: "The consumer you tried to insert is already in the database",
	NoConsumerFound:           "the supplied consumer id is not associated to any consumer",
}

var httpCodes = map[string]int{
	MissingAuthorizationInformation: http.StatusUnauthorized,
	InsufficientScope:               http.StatusForbidden,
	InternalError:                   http.StatusInternalServerError,
	UniqueConstraintViolation:       http.StatusConflict,
	NoConsumerFound:                 http.StatusNotFound,
}
