import sqlalchemy
import sqlalchemy.dialects
import geoalchemy2

import database

water_usage_meta_data = sqlalchemy.MetaData(schema="water_usage")
geodata_metadata = sqlalchemy.MetaData(schema="geodata")

usages = sqlalchemy.Table(
    "usages",
    water_usage_meta_data,
    sqlalchemy.Column("id", sqlalchemy.Integer, primary_key=True, autoincrement=True),
    sqlalchemy.Column("municipal", None, sqlalchemy.ForeignKey("geodata.nds_municipalities.id")),
    sqlalchemy.Column("consumer", None, sqlalchemy.ForeignKey("consumers.id")),
    sqlalchemy.Column("consumer_group", None, sqlalchemy.ForeignKey("consumer_group.id")),
    sqlalchemy.Column("year", sqlalchemy.Integer),
    sqlalchemy.Column("value", sqlalchemy.Numeric),
)

consumers = sqlalchemy.Table(
    "consumers",
    water_usage_meta_data,
    sqlalchemy.Column(
        "id", sqlalchemy.dialects.postgresql.UUID(as_uuid=True), primary_key=True, server_default="gen_random_uuid()"
    ),
    sqlalchemy.Column("name", sqlalchemy.Text),
    sqlalchemy.Column("location", geoalchemy2.Geometry("POINT", 4326)),
)

shapes = sqlalchemy.Table(
    "shapes",
    geodata_metadata,
    sqlalchemy.Column(
        "id", sqlalchemy.dialects.postgresql.UUID(as_uuid=True), primary_key=True, server_default="gen_random_uuid()"
    ),
    sqlalchemy.Column("name", sqlalchemy.Text),
    sqlalchemy.Column("key", sqlalchemy.Text),
    sqlalchemy.Column("geom", geoalchemy2.Geometry("MULTIPOLYGON", 4326)),
    sqlalchemy.Column("nuts_key", sqlalchemy.Text),
)


def initialize_tables():
    """
    Initialize the used tables
    """
    water_usage_meta_data.create_all(bind=database.engine)
