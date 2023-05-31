-- name: create-schema
CREATE SCHEMA IF NOT EXISTS consumers;

-- name: create-tables
CREATE TABLE IF NOT EXISTS consumers.consumers (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    location geometry(Point, 4326) NOT NULL,
    usage_type uuid,
    additional_properties jsonb
);

-- name: get-consumers-by-usage-id-area
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id IN (SELECT consumer FROM water_usage.usages WHERE usages.amount > $1)
AND
    id = any($2)
AND
    ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($3)), location);

-- name: get-consumers-by-usage-id
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id IN (SELECT consumer FROM water_usage.usages WHERE usages.amount > $1)
AND
    id = any($2);

-- name: get-consumers-by-usage-area
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id IN (SELECT consumer FROM water_usage.usages WHERE amount > $1)
AND
    ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($2)), location);

-- name: get-consumers-by-id-area
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id = any($1)
AND
    ST_CONTAINS((SELECT geom FROM geodata.shapes WHERE key = any($2)), location);

-- name: get-consumers-by-id
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id = any($1);

-- name: get-consumer-by-id
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
        id = $1::uuid;

-- name: get-consumers-by-area
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    st_contains((SELECT geom FROM geodata.shapes WHERE key = any($1)), location);

-- name: get-consumers-by-usage
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers
WHERE
    id IN (SELECT consumer FROM water_usage.usages WHERE amount > $1);

-- name: get-all-consumers
SELECT
    id,
    name,
    ST_ASGeoJSON(location) AS location,
    usage_type,
    additional_properties
FROM
    consumers.consumers;

-- name: insert-consumer
INSERT INTO
    consumers.consumers (name, location, usage_type, additional_properties)
VALUES
    ($1, st_makepoint($2, $3), $4, $5)
RETURNING id;

-- name: update-consumer-name
UPDATE
    consumers.consumers
SET
    name = $1
WHERE
    id = $2;

-- name: update-consumer-location
UPDATE
    consumers.consumers
SET
    location = st_makepoint($1, $2)
WHERE
    id = $3;

-- name: update-consumer-usage-type
UPDATE
    consumers.consumers
SET
    usage_type = $1
WHERE
    id = $2;

-- name: update-consumer-additional-properties
UPDATE
    consumers.consumers
SET
    additional_properties = $1::jsonb
WHERE
    id = $2;

-- name: delete-consumer
DELETE FROM
   consumers.consumers
WHERE
    id = $1;

-- name: get-consumer-type-id
SELECT
    id::text
FROM
    water_usage.usage_types
WHERE
    external_identifier = $1;

-- name: get-consumer-type-external-identifier
SELECT
    external_identifier
FROM
    water_usage.usage_types
WHERE
    id = $1;