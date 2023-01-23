package structs

import geoJson "github.com/paulmach/go.geojson"

// Consumer mirrors the database definition of a consumer
type Consumer struct {
	UUID     string           `json:"id"`
	Name     string           `json:"name"`
	Location geoJson.Geometry `json:"location"`
}
