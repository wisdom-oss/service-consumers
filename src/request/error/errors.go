// Package requestErrors contains all request errors which are directly handled by the handlers and are detected by
// the handlers. The request errors are identified by a constant value which also represents the error code
package requestErrors

import (
	"net/http"
)

const MissingAuthorizationInformation = "MISSING_AUTHORIZATION_INFORMATION"
const InsufficientScope = "INSUFFICIENT_SCOPE"
const InternalError = "INTERNAL_ERROR"

var titles = map[string]string{
	MissingAuthorizationInformation: "Unauthorized",
	InsufficientScope:               "Insufficient Scope",
	InternalError:                   "Internal Error",
}

var descriptions = map[string]string{
	MissingAuthorizationInformation: "The accessed resource requires authorization, " +
		"however the request did not contain valid authorization information. Please check the request",
	InsufficientScope: "The authorization was successful, " +
		"but the resource is protected by a scope which was not included in the authorization information",
	InternalError: "During the handling of the request an unexpected error occurred",
}

var httpCodes = map[string]int{
	MissingAuthorizationInformation: http.StatusUnauthorized,
	InsufficientScope:               http.StatusForbidden,
	InternalError:                   http.StatusInternalServerError,
}
