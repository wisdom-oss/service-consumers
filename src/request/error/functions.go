package requestErrors

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wisdom-oss/microservice-utils"
	"microservice/structs"
	"microservice/vars"
	"microservice/vars/globals"
	"net/http"
)

// RequestErrors contains all request errors which were read at the starup
// of the service
var RequestErrors map[string]structs.RequestError = make(map[string]structs.RequestError)

func GetRequestError(errorCode string) (*structs.ErrorResponse, error) {
	if !wisdomUtils.MapHasKey(RequestErrors, errorCode) {
		return nil, vars.ErrHttpErrorNotFound
	}
	// since the key exists in the request errors map get the request error
	requestError := RequestErrors[errorCode]
	// now get the text for the http error code
	httpText := http.StatusText(requestError.HttpCode)
	if httpText == "" {
		return &structs.ErrorResponse{
			HttpStatus:       requestError.HttpCode,
			HttpError:        "Unknown HTTP Code",
			ErrorCode:        fmt.Sprintf("%s.%s", globals.ServiceName, errorCode),
			ErrorTitle:       requestError.ErrorTitle,
			ErrorDescription: requestError.ErrorDescription,
		}, nil
	} else {
		return &structs.ErrorResponse{
			HttpStatus:       requestError.HttpCode,
			HttpError:        httpText,
			ErrorCode:        errorCode,
			ErrorTitle:       requestError.ErrorTitle,
			ErrorDescription: requestError.ErrorDescription,
		}, nil
	}
}

// WrapInternalError wraps an internal error and adds the error to the response
func WrapInternalError(err error) (*structs.ErrorResponse, error) {
	response, err := GetRequestError("INTERNAL_ERROR")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create internal error")
	}
	response.ErrorDescription += fmt.Sprintf("%s", err)
	return response, nil
}

// SendError takes the request error and sends it back to the client
func SendError(errorResponse *structs.ErrorResponse, w http.ResponseWriter) {
	// set the content-type header to json
	w.Header().Set("Content-Type", "text/json")
	// now send the http code to the client
	w.WriteHeader(errorResponse.HttpStatus)
	// now write out the response
	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		globals.HttpLogger.Err(err).Msg("unable to send error response")
	}
}
