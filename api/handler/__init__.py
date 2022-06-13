import http

import fastapi
import sqlalchemy.exc

import exceptions
import configuration


# %% Exception Handlers
async def handle_api_error(_: fastapi.requests.Request, exception: exceptions.APIException):
    content = {
        "httpCode": exception.http_status.value,
        "httpError": exception.http_status.phrase,
        "error": configuration.ServiceConfiguration().name + f".{exception.error_code}",
        "errorName": exception.error_title,
        "errorDescription": exception.error_description,
    }
    for key, value in content.items():
        if value is None:
            content.pop(key)
    return fastapi.responses.ORJSONResponse(status_code=exception.http_status.value, content=content)


async def handle_integrity_error(_: fastapi.requests.Request, _exception: sqlalchemy.exc.IntegrityError):
    content = {
        "httpCode": http.HTTPStatus.CONFLICT.value,
        "httpError": http.HTTPStatus.CONFLICT.phrase,
        "error": configuration.ServiceConfiguration().name + f".DUPLICATE_ENTRY",
        "errorName": "Constraint Violation",
        "errorDescription": "The resource you are trying to create already exists",
    }
    return fastapi.responses.ORJSONResponse(content=content, status_code=http.HTTPStatus.CONFLICT)


def handle_request_validation_error(_: fastapi.requests.Request, _exception: fastapi.exceptions.RequestValidationError):
    content = {
        "httpCode": http.HTTPStatus.BAD_REQUEST.value,
        "httpError": http.HTTPStatus.BAD_REQUEST.phrase,
        "error": configuration.ServiceConfiguration().name + f".BAD_REQUEST",
        "errorName": "Bad Request Parameters",
        "errorDescription": "The request did not contain all necessary parameters to be executed successfully",
    }
    return fastapi.responses.ORJSONResponse(content=content, status_code=http.HTTPStatus.BAD_REQUEST)
