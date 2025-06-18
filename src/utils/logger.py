import logging
import sys
from typing import Any
from pydantic import BaseModel

class LogConfig(BaseModel):
    LOGGER_NAME: str = "alya.io"
    LOG_FORMAT: str = "%(levelname)s | %(asctime)s | %(message)s"
    LOG_LEVEL: str = "INFO"

    # Logging Config
    version: int = 1
    disable_existing_loggers: bool = False
    formatters: dict[str, Any] = {
        "default": {
            "format": LOG_FORMAT,
            "datefmt": "%Y-%m-%d %H:%M:%S",
        },
    }

    handlers: dict[str, Any] = {
        "default": {
            "formatter": "default",
            "class": "logging.StreamHandler",
            "stream": "ext://sys.stdout",
        }
    }

    loggers: dict[str, Any] = {
        LOGGER_NAME: {"handlers": ["default"], "level": LOG_LEVEL}
    }


def setup_logging():
    import os
    from logging.config import dictConfig
    config = LogConfig()
    config.LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
    dictConfig(config.model_dump())
    return logging.getLogger(config.LOGGER_NAME)

logger = setup_logging()