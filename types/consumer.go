package types

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)
import "github.com/paulmach/go.geojson"

var ErrNoCoordinates = errors.New("no location set")
var ErrLocationNotTwoCoordinates = errors.New("array not containing two floats")

type Consumer struct {
	// ID contains the identifier of the consumer
	ID uuid.UUID `db:"id" json:"id"`

	// Name contains the name of the consumer
	Name string `db:"name" json:"name"`

	// Description contains a short and optional description of the consumer
	Description *string `db:"description" json:"description"`

	// Address contains a human-readable location of the consumer
	Address *string `db:"address" json:"address"`

	// Location contains the GeoJSON representation of the consumer's location
	// as a geometry
	Location *geojson.Geometry `db:"location" json:"location"`

	// UsageType contains the usage type that the consumer has been assigned to
	UsageType *uuid.UUID `db:"usage_type" json:"usageType"`

	// AdditionalProperties contain additional properties that further apply
	// to the consumer
	AdditionalProperties map[string]interface{} `db:"additional_properties" json:"additionalProperties"`
}

// UnmarshalJSON customizes the way this struct is populated when reading
// a JSON object.
// The main difference to the way the Consumer type is defined is that the
// location contains an array of two floats and not a GeoJSON representation
// of the location.
// This is done to make the creation of a new consumer easier
func (c *Consumer) UnmarshalJSON(src []byte) error {
	// this contains the type awaited as the incoming json object
	type incomingConsumer struct {
		Name                 string                 `json:"name"`
		Description          *string                `json:"description"`
		Address              *string                `json:"address"`
		Location             []float64              `json:"location"`
		UsageType            *uuid.UUID             `json:"usageType"`
		AdditionalProperties map[string]interface{} `json:"additional_properties"`
	}

	var iC incomingConsumer
	// now try to parse the incoming consumer
	err := json.Unmarshal(src, &iC)
	if err != nil {
		return err
	}

	var nC Consumer
	nC.Name = iC.Name
	nC.Description = iC.Description
	nC.Address = iC.Address
	nC.UsageType = iC.UsageType
	nC.AdditionalProperties = iC.AdditionalProperties

	// now create a location from the coordinates if they are available
	if iC.Location == nil {
		return ErrNoCoordinates
	}
	if len(iC.Location) != 2 {
		return ErrLocationNotTwoCoordinates
	}
	geom := geojson.NewPointGeometry(iC.Location)
	nC.Location = geom
	return nil
}
