import enum


class AMQPAction(str, enum.Enum):
    """The actions which are available for this service"""

    CHECK_TOKEN_SCOPE = "validate_token"
    """Check the scope of a token and return if the token is valid and has the scope"""

    ADD_SCOPE = "add_scope"
    """Add a scope to the authorization system"""

    EDIT_SCOPE = "edit_scope"
    """Edit a scope already in the authorization system"""

    CHECK_SCOPE = "check_scope"
    """Check if a scope is already present in the system"""


class TokenIntrospectionFailure(str, enum.Enum):
    """
    The reasons why a token introspection has failed and did not return that the token is valid
    """

    INVALID_TOKEN = "INVALID_TOKEN"
    """The token either has an invalid format or was not found in the database"""

    EXPIRED = "EXPIRED_TOKEN"
    """The tokens TTL as expired"""

    TOKEN_USED_TOO_EARLY = "USAGE_BEFORE_CREATION"
    """The token has been used before it's creation time"""

    NO_USER_ASSOCIATED = "NO_ASSOCIATED_USER"
    """The token has no user associated to it"""

    USER_DISABLED = "USER_DISABLED"
    """The user associated to the account has been disabled"""

    MISSING_PRIVILEGES = "MISSING_PRIVILEGES"
    """The scopes associated to this token are not matching the one required to access this endpoint"""
