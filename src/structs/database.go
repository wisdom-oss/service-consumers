package structs

import (
	geojson "github.com/paulmach/go.geojson"

	"microservice/vars/globals"
	"microservice/vars/globals/connections"
)
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

	Description string `db:"description"`

	Address string `db:"address"`
}

// ToConsumer converts a consumer entry stored in the database into an object
// which may be json-encoded later on
func (c DbConsumer) ToConsumer() (*Consumer, error) {
	var uuid string
	err := c.ID.AssignTo(&uuid)
	if err != nil {
		return nil, err
	}

	var usageType *string
	if c.UsageType.Status == pgtype.Present {
		var usageTypeUUID string
		err := c.UsageType.AssignTo(&usageTypeUUID)
		if err != nil {
			return nil, err
		}
		usageTypeRow, err := globals.Queries.QueryRow(connections.DbConnection, "get-consumer-type-external-identifier", usageTypeUUID)
		if err != nil {
			return nil, err
		}
		err = usageTypeRow.Scan(&usageType)
		if err != nil {
			return nil, err
		}
	} else {
		usageType = nil
	}

	var additionalProperties map[string]interface{}
	if c.AdditionalProperties.Status == pgtype.Present {
		err := c.AdditionalProperties.AssignTo(&additionalProperties)
		if err != nil {
			return nil, err
		}
	} else {
		additionalProperties = nil
	}

	return &Consumer{
		UUID:                 uuid,
		Name:                 c.Name,
		Location:             c.Location,
		Description:          c.Description,
		Address:              c.Address,
		UsageType:            usageType,
		AdditionalProperties: additionalProperties,
	}, nil
}
