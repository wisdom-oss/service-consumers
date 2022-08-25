// Package errors
// This package contains all errors the service may return in HTTP requests
// Those errors include unauthenticated calls and forbidden ones
package errors

import (
	"fmt"
	"net/http"

	"microservice/structs"
	"microservice/vars"
)

const UnauthorizedRequest = "UNAUTHORIZED_REQUEST"
const MissingScope = "SCOPE_MISSING"
const UnsupportedHTTPMethod = "UNSUPPORTED_METHOD"
const InvalidQueryParameter = "INVALID_QUERY_PARAMETER"
const DatabaseQueryError = "DATABASE_QUERY_ERROR"
const UnprocessableEntity = "UNPROCESSABLE_ENTITY"
const UniqueConstraintViolation = "UNIQUE_CONSTRAINT_VIOLATION"

var errorTitle = map[string]string{
	UnauthorizedRequest:       "Unauthorized Request",
	MissingScope:              "Forbidden",
	UnsupportedHTTPMethod:     "Unsupported HTTP Method",
	InvalidQueryParameter:     "Invalid Query Parameter",
	DatabaseQueryError:        "Database Query Error",
	UnprocessableEntity:       "Unprocessable Entity",
	UniqueConstraintViolation: "Unique Constraint Violation",
}

var errorDescription = map[string]string{
	UnauthorizedRequest: "The resource you tried to access requires authorization. Please check your request",
	MissingScope: "Yu tried to access a resource which is protected by a scope. " +
		"Your authorization information did not contain the required scope.",
	UnsupportedHTTPMethod: "The used HTTP method is not supported by this microservice. " +
		"Please check the documentation for further information",
	InvalidQueryParameter: "One or more parameters set for the request are not in a valid format. " +
		"Please check your request and read the API documentation.",
	DatabaseQueryError: "The microservice was unable to successfully execute the database query. " +
		"Please check the logs for more information",
	UnprocessableEntity: "The JSON object you sent to the service is not processable. Please check your request",
	UniqueConstraintViolation: "The object you are trying to create already exists in the database. " +
		"Please check your request and the documentation",
}

var httpStatus = map[string]int{
	UnauthorizedRequest:       http.StatusUnauthorized,
	MissingScope:              http.StatusForbidden,
	UnsupportedHTTPMethod:     http.StatusMethodNotAllowed,
	InvalidQueryParameter:     http.StatusBadRequest,
	DatabaseQueryError:        http.StatusInternalServerError,
	UnprocessableEntity:       http.StatusUnprocessableEntity,
	UniqueConstraintViolation: http.StatusConflict,
}

func NewRequestError(errorCode string) structs.RequestError {
	return structs.RequestError{
		HttpStatus:       httpStatus[errorCode],
		HttpError:        http.StatusText(httpStatus[errorCode]),
		ErrorCode:        fmt.Sprintf("%s.%s", vars.ServiceName, errorCode),
		ErrorTitle:       errorTitle[errorCode],
		ErrorDescription: errorDescription[errorCode],
	}
}
