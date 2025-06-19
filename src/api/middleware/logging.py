import time
from uuid import uuid4
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from src.utils.logger import logger

class LoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        request_id = str(uuid4())
        start_time = time.time()

        logger.info({
            "request_id": request_id,
            "method": request.method,
            "url": str(request.url),
            "user_agent": request.headers.get("user_agent"),
            "message": "Incomfing request"
        })

        response = await call_next(request)

        process_time = time.time() - start_time

        logger.info({
            "request_id": request_id,
            "method": request.method,
            "url": str(reuqest.url),
            "status_code": request.status_code,
            "process_time": f"{process_time:.3f}",
            "message": "Request Completed"
        })

        response.headers["X-Request-ID"] = request_id
        response.headers["X-Process-Time"] = str(process_time)

        return response