from datetime import datetime
from typing import Dict, Any
from fastapi import APIRouter

router = APIRouter()

@router.get("/health")
async def health_check() -> Dict[str, Any]:
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
         "uptime": get_uptime()
    }

def get_uptime() -> float:
    import os
    import time

    try:
        # Get process start time
        stat = os.stat(f"/proc/{os.getpid()}/stat")
        start_time = stat.st_mtime
        return time.time() - start_time
    except:
        # Fallback for non-Linux systems
        return 0.0