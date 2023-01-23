package requestErrors

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"microservice/structs"
	"microservice/utils"
	"microservice/vars"
	"net/http"
)

// BuildRequestError creates a RequestError which can be sent in case of an error which has been triggered.
//The code for an error is defined as constant in the `request/error/errors.go` file
func BuildRequestError(code string) (*structs.RequestError, error) {
	// Check if the error code is configured in the respective arrays
	if !utils.MapContainsKey(titles, code) ||
		!utils.MapContainsKey(descriptions, code) ||
		!utils.MapContainsKey(httpCodes, code) {
		return nil, vars.ErrHttpErrorNotFound
	}
	// Now build the request error struct and return it
	return &structs.RequestError{
		HttpStatus:       httpCodes[code],
		HttpError:        http.StatusText(httpCodes[code]),
		ErrorCode:        fmt.Sprintf("%s.%s", vars.ServiceName, code),
		ErrorTitle:       titles[code],
		ErrorDescription: descriptions[code],
	}, nil
}

// RespondWithRequestError responds with an already built request error
func RespondWithRequestError(requestError *structs.RequestError, responseWriter http.ResponseWriter) {
	// Set the content type of the response to "text/json"
	responseWriter.Header().Set("Content-Type", "text/json")
	// Write the http status of the request error to the response
	responseWriter.WriteHeader(requestError.HttpStatus)
	// Now encode the request error to json and write it to the response
	encodingError := json.NewEncoder(responseWriter).Encode(requestError)
	if encodingError != nil {
		log.WithField("package", "request/error").WithError(encodingError).Error(
			"unable to encode the struct into json")
	}
}

// RespondWithInternalError creates an Internal Server Error which contains the error thrown in the microservice as
// part of the error description
func RespondWithInternalError(reason error, responseWriter http.ResponseWriter) {
	requestError, _ := BuildRequestError(InternalError)
	// Put the reason into the description of the error
	requestError.ErrorDescription = fmt.Sprintf("%s: %s", requestError.ErrorDescription, reason)
	// Now send the response
	RespondWithRequestError(requestError, responseWriter)
}
