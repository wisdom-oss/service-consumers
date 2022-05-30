import logging

import sqlalchemy.engine

import configuration

_logger = logging.getLogger(__name__)

_settings = configuration.DatabaseConfiguration()

engine = sqlalchemy.engine.create_engine(_settings.dsn, pool_recycle=90)
