package structs

import geoJson "github.com/paulmach/go.geojson"

// ScopeInformation contains the information about the scope for this service
type ScopeInformation struct {
	JSONSchema       string `json:"$schema"`
	ScopeName        string `json:"name"`
	ScopeDescription string `json:"description"`
	ScopeValue       string `json:"scopeStringValue"`
}

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
	UUID     string           `json:"id"`
	Name     string           `json:"name"`
	Location geoJson.Geometry `json:"location"`
}

// IncomingConsumer contains the data which is for updating or creating a consumer
type IncomingConsumer struct {
	Name      *string  `json:"name"`
	Latitude  *float64 `json:"lat"`
	Longitude *float64 `json:"long"`
}
