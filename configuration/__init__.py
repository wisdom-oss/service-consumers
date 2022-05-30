import typing

import pydantic


class ServiceConfiguration(pydantic.BaseSettings):

    name: str = pydantic.Field(
        default=...,
        title="Microservice Name",
        description="The name of the service which will be used during the registration at the service registry",
        env="CONFIG_SERVICE_NAME",
        alias="CONFIG_SERVICE_NAME",
    )
    """
    Microservice Name

    The name of the microservice. The name will be used to identify this service and it's instances at the service
    registry
    """

    http_port: int = pydantic.Field(
        default=5000,
        title="HTTP Port",
        description="The http port which will be bound by the service in the container",
        env="CONFIG_HTTP_PORT",
        alias="CONFIG_HTTP_PORT",
    )
    """
    HTTP Port

    The http port which will be bound by the service in the container
    """

    logging_level: str = pydantic.Field(
        default="INFO",
        title="Logging Level",
        description="The level which is used for the root logger, which will display messages from this and levels "
        "above",
        env="CONFIG_LOGGING_LEVEL",
        alias="CONFIG_LOGGING_LEVEL",
    )
    """
    Logging Level

    The level which is used for the root logger. The root logger will display messages from this level and levels
    above this one.
    """

    class Config:
        env_file = ".env"


class AMQPConfiguration(pydantic.BaseSettings):

    dsn: pydantic.AmqpDsn = pydantic.Field(
        default=...,
        title="AMQP Data Source Name",
        description="The data source name pointing to an installation of a RabbitMQ message broker",
        env="CONFIG_AMQP_DSN",
        alias="CONFIG_AMQP_DSN",
    )
    """
    AMQP Data Source Name

    The data source name pointing to an installation of the RabbitMQ message broker
    """

    exchange: typing.Optional[str] = pydantic.Field(
        default=None,
        title="AMQP Send Exchange",
        description="The exchange to which this service will send messages",
        env="CONFIG_AMQP_EXCHANGE",
        alias="CONFIG_AMQP_EXCHANGE",
    )
    """
    AMQP Send Exchange

    The exchange to which this service will send the messages
    """

    authorization_exchange: typing.Optional[str] = pydantic.Field(
        default="authorization-service",
        title="AMQP Authorization Exchange",
        description="The exchange to which the authorization service listens to",
        env="CONFIG_AMQP_AUTHORIZATION_EXCHANGE",
        alias="CONFIG_AMQP_AUTHORIZATION_EXCHANGE",
    )
    """
    AMQP Authorization Service

    The exchange to which this service will send the messages related to authorizing users and requests
    """

    class Config:
        env_file = ".env"


class SecurityConfiguration(pydantic.BaseSettings):

    scope_string_value: typing.Optional[str] = pydantic.Field(
        default=None,
        title="Required Scope String value",
        description="The scope string value of the scope which is required to access the service",
        env="CONFIG_SECURITY_SCOPE",
        alias="CONFIG_SECURITY_SCOPE",
    )
    """
    Required Scope String Value

    The scope string value of the scope which is required to access this service. If no value is set the access to
    the services routes are unprotected
    """

    class Config:
        env_file = ".env"


class ServiceRegistryConfiguration(pydantic.BaseSettings):
    """Settings which will influence the connection to the service registry"""

    host: str = pydantic.Field(default=..., alias="CONFIG_SERVICE_REGISTRY_HOST", env="CONFIG_SERVICE_REGISTRY_HOST")
    """
    Eureka Service Registry Host

    The host on which the eureka service registry is running on.
    """

    port: int = pydantic.Field(default=8761, alias="CONFIG_SERVICE_REGISTRY_PORT", env="CONFIG_SERVICE_REGISTRY_PORT")
    """
    Eureka Service Registry Port

    The port on which the eureka service registry is running on.
    """

    class Config:
        env_file = ".env"


class DatabaseConfiguration(pydantic.BaseSettings):
    """Settings which are related to the database connection"""

    dsn: pydantic.PostgresDsn = pydantic.Field(default=..., alias="CONFIG_DB_DSN", env="CONFIG_DB_DSN")
    """
    PostgreSQL data source name

    The data source name (expressed as URI) pointing to the installation of the used postgresql database
    """

    class Config:
        env_file = ".env"
