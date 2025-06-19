import os
import signal
import sys
from contextlib import asynccontextmanager
import uvicorn
from dotenv import load_dotenv

from src.api.app import create_app
from src.utils.logger import logger 

load_dotenv()

@asynccontextmanager
async def lifespan(app):
    # Startup
    logger.info({"message": "Starting up...", "port": os.getenv("PORT", 8000)})
    yield
    # Shutdown
    logger.info("Shutting down...")

app = create_app(lifespan=lifespan)

def signal_handler(sig, frame):
    logger.info(f"Received signal {sig}")
    sys.exit(0)

if __name__ == "__main__":
    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    port = int(os.getenv("PORT", 8000))
    host = os.getenv("HOST", "0.0.0.0")
    reload = os.getenv("ENV", "development") == "development"

    logger.info({
        "message": "Server started successfully",
        "port": port,
        "host": host,
        "environment": os.getenv("ENV", "development")
    })

    uvicorn.run(
        "src.main:app",
        host=host,
        port=port,
        reload=reload,
        log_config=None
    )

