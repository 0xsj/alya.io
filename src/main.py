import os
import signal
import sys
from contextlib import asynccontextmanager
import uvicorn
from dotenv import load_dotenv

from src.api.app import create_app
from src.utils.logger import logger 