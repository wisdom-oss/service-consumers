# WISdoM OSS - Consumer Management Service

This microservice deploys an interface to manage the consumers present in the
database. The documentation of the RESTful-API is stored in the
[openapi.yaml](openapi.yaml) file in the root directory of the repository.

## Deployment
This service is included in the standard installation of the project since it is
one of the core services.

## Access
This microservice will register itself to the API gateway of the installation.
Therefore, you can access this service on the following path, if the path is
not changed manually: `https://<your-server-address>/api/consumers`

## Configuration

The microservice is configurable via the following environment variables:
- `CONFIG_LOGGING_LEVEL` &#8594; Set the logging verbosity [optional, default `INFO`]
- `CONFIG_API_GATEWAY_HOST` &#8594; Set the host on which the API Gateway runs on **[required]**
- `CONFIG_API_GATEWAY_ADMIN_PORT` &#8594; Set the port on which the API Gateway listens on **[required]**
- `CONFIG_API_GATEWAY_SERVICE_PATH` &#8594; Set the path under which the service shall be reachable. _Do not prepend the path with `/api`. Only set the last part of the desired path_ **[required]**
- `CONFIG_HTTP_LISTEN_PORT` &#8594; The port on which the built-in webserver will listen on [optional, default `8000`]
- `CONFIG_SCOPE_FILE_PATH` &#8594; The location where the scope definition file is stored [optional, default `/microservice/res/scope.json]

