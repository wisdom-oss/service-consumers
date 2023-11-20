package structs

import geoJson "github.com/paulmach/go.geojson"

// ErrorResponse contains all information about an error which shall be sent back to the client
type ErrorResponse struct {
	HttpStatus       int    `json:"httpCode"`
	HttpError        string `json:"httpError"`
	ErrorCode        string `json:"error"`
	ErrorTitle       string `json:"errorName"`
	ErrorDescription string `json:"errorDescription"`
}

// Consumer is a struct which is used to serialize the query outputs from the
// database
type Consumer struct {
	UUID                 string                 `json:"id"`
	Name                 string                 `json:"name"`
	Location             geoJson.Geometry       `json:"location"`
	UsageType            *string                `json:"usageType"`
	Description          string                 `json:"description"`
	Address              string                 `json:"address"`
	AdditionalProperties map[string]interface{} `json:"additionalProperties"`
}

// IncomingConsumer contains the data which is for updating or creating a consumer
type IncomingConsumer struct {
	Name                 *string                 `json:"name"`
	Coordinates          *[2]float64             `json:"coordinates"`
	UsageType            *string                 `json:"usageType"`
	AdditionalProperties *map[string]interface{} `json:"additionalProperties"`
}
