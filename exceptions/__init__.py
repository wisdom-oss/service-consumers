import http
import typing


class APIException(Exception):
    """
    An exception for returning any error that happened in the API
    """

    def __init__(
        self,
        error_code: str,
        error_title: typing.Optional[str] = None,
        error_description: typing.Optional[str] = None,
        http_status: typing.Union[http.HTTPStatus, int] = http.HTTPStatus.INTERNAL_SERVER_ERROR,
    ):
        """
        Create a new API exception

        :param error_code: The error code of the exception
        :type error_code: str
        :param error_title: The title of the exception
        :type error_title: str
        :param error_description: The description of the exceptions
        :type error_description: str
        :param http_status: The HTTP Status that will be sent back
        :type http_status: http.HTTPStatus
        """
        super().__init__()
        self.error_code = error_code
        self.error_title = error_title
        self.error_description = error_description
        self.http_status = http_status
