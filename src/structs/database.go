package structs

import geojson "github.com/paulmach/go.geojson"
import "github.com/jackc/pgtype"

// DbConsumer reflects a consumer stored in the database. It may
// be converted to a json-encodable output by using the ToConsumer() function
// on the given instance
type DbConsumer struct {
	// ID contains the database-generated UUID for this consumer
	ID pgtype.UUID `db:"id"`

	// Name contains the name of the consumer
	Name string `db:"name"`

	// Location contains the GeoJSON representation of the consumers location
	Location geojson.Geometry `db:"location"`

	// UsageType contains the UUID of the usage type that was assigned to the
	// consumer
	UsageType pgtype.UUID `db:"usage_type"`

	// AdditionalProperties contains an optional key/value map allowing to add
	// additional properties to a consumer
	AdditionalProperties pgtype.JSONB `db:"additional_properties"`
}

// ToConsumer converts a consumer entry stored in the database into an object
// which may be json-encoded later on
func (c DbConsumer) ToConsumer() Consumer {
	var uuid string
	c.ID.AssignTo(&uuid)

	var usageType *string
	if c.UsageType.Status == pgtype.Present {
		c.UsageType.AssignTo(usageType)
	} else {
		usageType = nil
	}

	var additionalProperties map[string]interface{}
	if c.AdditionalProperties.Status == pgtype.Present {
		c.AdditionalProperties.AssignTo(&additionalProperties)
	} else {
		additionalProperties = nil
	}

	return Consumer{
		UUID:                 uuid,
		Name:                 c.Name,
		Location:             c.Location,
		UsageType:            usageType,
		AdditionalProperties: additionalProperties,
	}
}
