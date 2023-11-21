-- name: get-consumers
SELECT
    id,
    name,
    description,
    address,
    ST_AsGeoJSON(location) as location,
    usage_type,
    additional_properties
FROM
    consumers.consumers;

-- name: filter-ids
id = any($1);

-- name: filter-usage-amount
id IN (SELECT consumer FROM water_usage.usages WHERE consumer IS NOT NULL AND usages.amount > $1);

-- name: filter-location
ST_CONTAINS(ST_UNION(ARRAY((SELECT geom FROM geodata.shapes WHERE key = any($1)))), location);
