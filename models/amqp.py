import pydantic

from . import BaseModel as _BaseModel


class TokenIntrospectionRequest(_BaseModel):
    """
    The data model describing how a token introspection request will look like
    """

    action: str = pydantic.Field(default="validate_token", alias="action")
    """The action that shall be executed on the authorization server"""

    bearer_token: str = pydantic.Field(default=..., alias="token")
    """The Bearer token that has been extracted and now shall be validated"""

    scope: str = pydantic.Field(default=..., alias="scope")
    """The scope which needs to be in the tokens scope to pass the introspection"""


class CreateScopeRequest(_BaseModel):
    """
    The data model describing how a scope creation request will look like
    """

    action: str = pydantic.Field(default="add_scope")

    name: str = pydantic.Field(default=..., alias="name")
    """The name of the new scope"""

    description: str = pydantic.Field(default=..., alias="description")
    """The description of the new scope"""

    value: str = pydantic.Field(default=..., alias="value")
    """String which will identify the scope"""

    @pydantic.validator("value")
    def check_scope_value_for_whitespaces(cls, v: str):
        if " " in v:
            raise ValueError("The scope value may not contain any whitespaces")
        return v


class CheckScopeRequest(_BaseModel):

    action: str = pydantic.Field(default="check_scope", alias="action")

    value: str = pydantic.Field(default=..., alias="scope")
    """The value of the scope that shall tested for existence"""
