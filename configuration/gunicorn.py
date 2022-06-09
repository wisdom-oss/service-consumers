import asyncio
import logging
import os
import sys
import typing

import amqp_rpc_client
import orjson
import py_eureka_client.eureka_client
import pydantic

import configuration
import models.amqp
import models.internal
import tools

bind = f"0.0.0.0:{configuration.ServiceConfiguration().http_port}"
workers = 1
limit_request_line = 0
limit_request_fields = 0
limit_request_field_size = 0
worker_class = "uvicorn.workers.UvicornWorker"
max_requests = 0
timeout = 0

_service_registry_client: typing.Optional[py_eureka_client.eureka_client.EurekaClient] = None


def on_starting(server):
    _service_configuration = configuration.ServiceConfiguration()
    logging.basicConfig(
        format="[%(asctime)s] [%(process)d] [%(levelname)s] %(message)s", datefmt="%Y-%m-%d %H:%M:%S %z"
    )
    # %% Validate the Service Registry settings and reachability
    try:
        _service_registry_configuration = configuration.ServiceRegistryConfiguration()
    except pydantic.ValidationError:
        logging.critical(
            "Unable to read the service registry related settings. Please refer to "
            "the documentation for further instructions: "
            "SERVICE_REGISTRY_SETTINGS_INVALID"
        )
        sys.exit(1)
    logging.info("Checking the connection to the service registry")
    _registry_available = asyncio.run(
        tools.is_host_available(
            host=_service_registry_configuration.host, port=_service_registry_configuration.port, timeout=10
        )
    )
    if not _registry_available:
        logging.critical(
            "The service registry is not reachable. Therefore, this service is unable to register "
            "itself at the service registry and it is not callable"
        )
        sys.exit(2)
    # %% Set up the service registry client
    global _service_registry_client
    _service_registry_client = py_eureka_client.eureka_client.EurekaClient(
        eureka_server=f"http://{_service_registry_configuration.host}:{_service_registry_configuration.port}/",
        app_name=_service_configuration.name,
        instance_port=_service_configuration.http_port,
        should_register=True,
        should_discover=False,
        renewal_interval_in_secs=5,
        duration_in_secs=30,
    )
    _service_registry_client.start()
    _service_registry_client.status_update("STARTING")
    # %% Validate the AMQP configuration and message broker reachability
    try:
        _amqp_configuration = configuration.AMQPConfiguration()
    except pydantic.ValidationError:
        logging.critical(
            "Unable to read the service registry related settings. Please refer to "
            "the documentation for further instructions: "
            "AMQP_CONFIGURATION_INVALID"
        )
        _service_registry_client.stop()
        sys.exit(2)
    _amqp_configuration.dsn.port = 5672 if _amqp_configuration.dsn.port is None else int(_amqp_configuration.dsn.port)
    _message_broker_reachable = asyncio.run(
        tools.is_host_available(_amqp_configuration.dsn.host, _amqp_configuration.dsn.port)
    )
    if not _message_broker_reachable:
        logging.error(
            f"The message broker is currently not reachable on {_amqp_configuration.dsn.host}:"
            f"{_amqp_configuration.dsn.port}"
        )
        _service_registry_client.stop()
        sys.exit(2)
    # %% Check if the configured service scope is available
    # Create an amqp client
    _amqp_client = amqp_rpc_client.Client(amqp_dsn=_amqp_configuration.dsn, mute_pika=True)
    service_scope = models.internal.ServiceScope.parse_file("./configuration/scope.json")
    # Query if the scope already exists
    _scope_check_request = models.amqp.CheckScopeRequest(value=service_scope.value)
    _scope_check_request_id = _amqp_client.send(
        _scope_check_request.json(by_alias=True), _amqp_configuration.authorization_exchange, "authorization-service"
    )
    _scope_check_response_bytes = _amqp_client.await_response(_scope_check_request_id)
    _scope_check_response: dict = orjson.loads(_scope_check_response_bytes)
    # Check if the scope check response contains any of the known error keys
    if set(_scope_check_response.keys()).issubset({"httpCode", "httpError", "error", "errorName", "errorDescription"}):
        # Since the scope check response contains an error request the scope to be created
        _scope_create_request = models.amqp.CreateScopeRequest(
            name=service_scope.name, description=service_scope.description, value=service_scope.value
        )
        _scope_create_request_id = _amqp_client.send(
            _scope_create_request.json(), _amqp_configuration.authorization_exchange, "authorization-service"
        )
        _scope_create_response_bytes = _amqp_client.await_response(_scope_create_request_id)
        _scope_create_response: dict = orjson.loads(_scope_create_response_bytes)
        if set(_scope_create_response.keys()).issubset(
            {"httpCode", "httpError", "error", "errorName", "errorDescription"}
        ):
            logging.critical(
                "Unable to create the scope which shall be used by the service:\n%s", _scope_create_response
            )
            _service_registry_client.stop()
            sys.exit(3)
        logging.info("Successfully created the scope that shall be used by this service")
    # Set the value for the used security scope internally
    os.environ["CONFIG_SECURITY_SCOPE"] = service_scope.value
    # %% Validate the database settings and reachability
    try:
        _database_configuration = configuration.DatabaseConfiguration()
    except pydantic.ValidationError:
        logging.critical(
            "Unable to read the service registry related settings. Please refer to "
            "the documentation for further instructions: "
            "DATABASE_CONFIGURATION_INVALID"
        )
        _service_registry_client.stop()
        sys.exit(1)
    logging.info("Checking the connection to the database")
    _database_configuration.dsn.port = (
        5432 if _database_configuration.dsn.port is None else int(_database_configuration.dsn.port)
    )
    _database_available = asyncio.run(
        tools.is_host_available(
            host=_database_configuration.dsn.host, port=_database_configuration.dsn.port, timeout=10
        )
    )
    if not _database_available:
        logging.critical(
            "The database is not available. Since this service requires an access to the database the service will "
            "not start"
        )
        _service_registry_client.stop()
        sys.exit(2)


def when_ready(server):
    _service_registry_client.status_update("UP")


def on_exit(server):
    _service_registry_client.stop()
