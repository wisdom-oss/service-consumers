package structs

import geoJson "github.com/paulmach/go.geojson"

type ScopeInformation struct {
	JSONSchema       string `json:"$schema"`
	ScopeName        string `json:"name"`
	ScopeDescription string `json:"description"`
	ScopeValue       string `json:"scopeStringValue"`
}

type RequestError struct {
	HttpStatus       int    `json:"httpCode"`
	HttpError        string `json:"httpError"`
	ErrorCode        string `json:"error"`
	ErrorTitle       string `json:"errorName"`
	ErrorDescription string `json:"errorDescription"`
}

/*
Consumer

The consumer holds the following information:
	- UUID: The internal and external id for the consumer
	- Name: The name of the consumer
	- Location: A GeoJSON entity containing the location of the consumer
*/
type Consumer struct {
	UUID     string           `json:"id"`
	Name     string           `json:"name"`
	Location geoJson.Geometry `json:"location"`
}
