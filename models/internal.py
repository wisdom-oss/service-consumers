import typing

import pydantic

import enums
from . import BaseModel as __BaseModel


class ServiceScope(__BaseModel):

    name: str = pydantic.Field(default=..., alias="name")
    """The name of the scope this service uses"""

    description: str = pydantic.Field(default=..., alias="description")
    """The description of the scope this service uses"""

    value: str = pydantic.Field(default=..., alias="scopeStringValue")
    """The string by which this scope is identifiable in a OAuth 2.0 scope string"""


class UserAccount(__BaseModel):
    id: int = pydantic.Field(default=..., title="Internal Account ID")
    """Internal Account ID"""

    first_name: str = pydantic.Field(default=..., title="First Name")
    """The first name of the user who is the owner of the account"""

    last_name: str = pydantic.Field(default=..., title="Last Name")
    """The last name of the user who is the owner of the account"""

    username: str = pydantic.Field(default=..., title="Username")
    """The username of the account"""


class TokenIntrospection(__BaseModel):

    active: bool = pydantic.Field(default=...)

    reason: typing.Optional[enums.TokenIntrospectionFailure] = pydantic.Field(default=None)

    scope: typing.Optional[typing.Union[str, list[str]]] = pydantic.Field(default=None)

    token_type: typing.Optional[str] = pydantic.Field(default=None)

    expires_at: typing.Optional[int] = pydantic.Field(default=None, alias="exp")

    created_at: typing.Optional[int] = pydantic.Field(default=None, alias="iat")

    user: typing.Optional[UserAccount] = pydantic.Field(default=None, alias="user")

    @pydantic.validator("scope")
    def convert_scope_list_to_string(cls, v):
        if type(v) is list:
            return v
        elif type(v) is str:
            return v.split()
        elif v is None:
            return v
        else:
            raise TypeError("The scope parameter only accepts lists or strings")
