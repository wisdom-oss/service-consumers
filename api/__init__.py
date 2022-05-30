"""Package containing the code which will be the API later on"""
import datetime
import email.utils
import hashlib
import http
import typing
import uuid

import amqp_rpc_client
import fastapi
import geoalchemy2.functions
import py_eureka_client.eureka_client
import pytz as pytz
import sqlalchemy.dialects
import sqlalchemy.exc
import ujson

import api.handler
import configuration
import database
import database.tables
import exceptions
import models.internal
import tools
from api import security

# %% Global Clients
_amqp_client: typing.Optional[amqp_rpc_client.Client] = None
_service_registry_client: typing.Optional[py_eureka_client.eureka_client.EurekaClient] = None

# %% API Setup
service = fastapi.FastAPI()
service.add_exception_handler(exceptions.APIException, api.handler.handle_api_error)
service.add_exception_handler(fastapi.exceptions.RequestValidationError, api.handler.handle_request_validation_error)
service.add_exception_handler(sqlalchemy.exc.IntegrityError, api.handler.handle_integrity_error)

# %% Configurations
_security_configuration = configuration.SecurityConfiguration()
if _security_configuration.scope_string_value is None:
    scope = models.internal.ServiceScope.parse_file("./configuration/scope.json")
    _security_configuration.scope_string_value = scope.value


# %% Middlewares
@service.middleware("http")
async def etag_comparison(request: fastapi.Request, call_next):
    """
    A middleware which will hash the request path and all parameters transferred to this
    microservice and will check if the hash matches the one of the ETag which was sent to the
    microservice. Furthermore, it will take the generated hash and append it to the response to
    allow caching

    :param request: The incoming request
    :type request: fastapi.Request
    :param call_next: The next call after this middleware
    :type call_next: callable
    :return: The result of the next call after this middle ware
    :rtype: fastapi.Response
    """
    # Access all parameters used for creating the hash
    path = request.url.path
    query_parameter = dict(request.query_params)
    content_type = request.headers.get("Content-Type", "text/plain")
    # Now iterate through all query parameters and make sure they are sorted if they are lists
    for key, value in dict(query_parameter).items():
        # Now check if the value is a list
        if isinstance(value, list):
            query_parameter[key] = sorted(value)

    query_dict = {
        "request_path": path,
        "request_query_parameter": query_parameter,
    }
    query_data = ujson.dumps(query_dict, ensure_ascii=False, sort_keys=True)
    # Now create a hashsum of the query data
    query_hash = hashlib.sha3_256(query_data.encode("utf-8")).hexdigest()
    # Now access the headers of the request and check for the If-None-Match Header
    if_none_match_value = request.headers.get("If-None-Match")
    if_modified_since_value = request.headers.get("If-Modified-Since")
    if if_modified_since_value is None:
        if_modified_since_value = datetime.datetime.fromtimestamp(0, tz=pytz.UTC)
    else:
        if_modified_since_value = email.utils.parsedate_to_datetime(if_modified_since_value)
    # Get the last update of the schema from which the service gets its data from
    # TODO: Set your schema name here
    last_database_modification = tools.get_last_schema_update("<<your-schema-name-here>>", database.engine)
    data_changed = if_modified_since_value < last_database_modification
    if query_hash == if_none_match_value and not data_changed:
        return fastapi.Response(status_code=304, headers={"ETag": f"{query_hash}"})
    else:
        response: fastapi.Response = await call_next(request)
        response.headers.append("ETag", f"{query_hash}")
        response.headers.append("Last-Modified", email.utils.format_datetime(last_database_modification))
        return response


# %% Routes
@service.get("/")
def get_all_consumers(
    usage_above: typing.Optional[int] = fastapi.Query(default=None),
    consumer_ids: list[uuid.UUID] = fastapi.Query(default=None, alias="id"),
    in_area: list[str] = fastapi.Query(default=None, alias="in"),
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    # %% Columns
    query_columns = [
        database.tables.consumers.c.id,
        database.tables.consumers.c.name,
        sqlalchemy.cast(
            geoalchemy2.functions.ST_AsGeoJSON(database.tables.consumers.c.location),
            sqlalchemy.dialects.postgresql.JSONB,
        ).label("geojson"),
    ]

    # %% Filter Building
    parameter_available = (usage_above is not None, consumer_ids is not None, in_area is not None)
    match parameter_available:
        case (True, True, True):
            query_filter = sqlalchemy.and_(
                database.tables.consumers.c.id.in_(
                    sqlalchemy.select([database.tables.usages.c.consumer], database.tables.usages.c.value > usage_above)
                ),
                database.tables.consumers.c.id.in_(consumer_ids),
                geoalchemy2.functions.ST_Contains(
                    sqlalchemy.select([database.tables.shapes.c.geom], database.tables.shapes.c.key.in_(in_area)),
                    database.tables.consumers.c.location,
                ),
            )
        case (True, True, False):
            query_filter = sqlalchemy.and_(
                database.tables.consumers.c.id.in_(
                    sqlalchemy.select([database.tables.usages.c.consumer], database.tables.usages.c.value > usage_above)
                ),
                database.tables.consumers.c.id.in_(consumer_ids),
            )
        case (True, False, True):
            query_filter = sqlalchemy.and_(
                database.tables.consumers.c.id.in_(
                    sqlalchemy.select([database.tables.usages.c.consumer], database.tables.usages.c.value > usage_above)
                ),
                geoalchemy2.functions.ST_Contains(
                    sqlalchemy.select([database.tables.shapes.c.geom], database.tables.shapes.c.key.in_(in_area)),
                    database.tables.consumers.c.location,
                ),
            )
        case (False, True, True):
            query_filter = sqlalchemy.and_(
                database.tables.consumers.c.id.in_(consumer_ids),
                geoalchemy2.functions.ST_Contains(
                    sqlalchemy.select([database.tables.shapes.c.geom], database.tables.shapes.c.key.in_(in_area)),
                    database.tables.consumers.c.location,
                ),
            )
        case (True, False, False):
            query_filter = database.tables.consumers.c.id.in_(
                sqlalchemy.select([database.tables.usages.c.consumer], database.tables.usages.c.value > usage_above)
            )
        case (False, True, False):
            query_filter = database.tables.consumers.c.id.in_(consumer_ids)
        case (False, False, True):
            query_filter = geoalchemy2.functions.ST_Contains(
                sqlalchemy.select([database.tables.shapes.c.geom], database.tables.shapes.c.key.in_(in_area)),
                database.tables.consumers.c.location,
            )
        case (False, False, False):
            query_filter = None
        case _:
            query_filter = None
    # %% Query Build
    consumer_query = sqlalchemy.select(query_columns, query_filter)
    consumers = database.engine.execute(consumer_query).all()
    if len(consumers) == 0:
        return fastapi.Response(status_code=204)
    return consumers


@service.put("/", status_code=http.HTTPStatus.CREATED)
def create_new_consumer(
    consumer_data: models.NewConsumerData = fastapi.Body(default=...),
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    # Create the point geometry
    point_ewkt = f"POINT({consumer_data.latitude} {consumer_data.longitude})"
    insert_query = sqlalchemy.insert(database.tables.consumers).values(name=consumer_data.name, location=point_ewkt)
    insert_result = database.engine.execute(insert_query).all()
    consumer_query = sqlalchemy.select(
        [
            database.tables.consumers.c.id,
            database.tables.consumers.c.name,
            sqlalchemy.cast(
                geoalchemy2.functions.ST_AsGeoJSON(database.tables.consumers.c.location),
                sqlalchemy.dialects.postgresql.JSONB,
            ).label("geojson"),
        ],
        database.tables.consumers.c.id == insert_result.inserted_primary_key[0],
    )
    consumer = database.engine.execute(consumer_query).all()
    return consumer


@service.patch("/{consumer_id}")
async def update_consumer(
    consumer_id: uuid.UUID,
    consumer_data: models.ConsumerUpdateData = fastapi.Body(default=...),
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    consumer_query = sqlalchemy.select([database.tables.consumers], database.tables.consumers.c.id == consumer_id)
    consumer = database.engine.execute(consumer_query).first()
    update_consumer_query = (
        sqlalchemy.update(database.tables.consumers)
        .where(database.tables.consumers.c.id == consumer_id)
        .values(
            name=consumer_data.name if consumer_data.name is not None else consumer[1],
            location=consumer[2]
            if consumer_data.latitude is None or consumer_data.longitude is None
            else f"POINT({consumer_data.latitude} {consumer_data.longitude})",
        )
    )
    database.engine.execute(update_consumer_query)
    consumer_query = sqlalchemy.select(
        [
            database.tables.consumers.c.id,
            database.tables.consumers.c.name,
            sqlalchemy.cast(
                geoalchemy2.functions.ST_AsGeoJSON(database.tables.consumers.c.location),
                sqlalchemy.dialects.postgresql.JSONB,
            ).label("geojson"),
        ],
        database.tables.consumers.c.id == consumer_id,
    )
    consumer = database.engine.execute(consumer_query).all()
    return consumer


@service.delete("/{consumer_id}", status_code=204)
async def delete_consumer(
    consumer_id: uuid.UUID,
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    consumer_delete = sqlalchemy.delete(database.tables.consumers).where(database.tables.consumers.c.id == consumer_id)
    database.engine.execute(consumer_delete)
    return fastapi.Response(status_code=204)
