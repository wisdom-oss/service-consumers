import pydantic


class BaseModel(pydantic.BaseModel):
    """The base model for all other models which has some preconfigured configuration"""

    class Config:
        """The configuration that all models will inherit if it bases itself on this BaseModel"""

        extra = pydantic.Extra.allow
        """Allow extra attributes to be assigned to the model"""

        allow_population_by_field_name = True
        """Allow fields to be populated by their name and alias"""

        smart_union = True
        """Check all types of a Union to prevent converting types"""


class NewConsumerData(BaseModel):

    name: str = pydantic.Field(default=..., alias="name")
    """The name of the new consumer"""

    latitude: float = pydantic.Field(default=..., alias="lat")
    """The latitude of the consumers location"""

    longitude: float = pydantic.Field(default=..., alias="long")
    """The longitude of the consumers location"""


class ConsumerUpdateData(BaseModel):

    name: str = pydantic.Field(default=None, alias="name")
    """The name of the new consumer"""

    latitude: float = pydantic.Field(default=None, alias="lat")
    """The latitude of the consumers location"""

    longitude: float = pydantic.Field(default=None, alias="long")
    """The longitude of the consumers location"""
