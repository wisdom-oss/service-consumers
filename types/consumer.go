package types

import "github.com/google/uuid"
import "github.com/paulmach/go.geojson"

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
	Location geojson.Geometry `db:"location" json:"location"`

	// UsageType contains the usage type that the consumer has been assigned to
	UsageType *uuid.UUID `db:"usage_type" json:"usageType"`

	// AdditionalProperties contain additional properties that further apply
	// to the consumer
	AdditionalProperties map[string]interface{} `db:"additional_properties" json:"additionalProperties"`
}
