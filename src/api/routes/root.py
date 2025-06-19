from typing import Dict
from fastapi import APIRouter
from src.core.config import settings

router = APIRouter()

@router.get("/")
async def root() -> Dict[str, str]:
    return {
        "message": "Welcome to alya.io",
        "version": settings.VERSION
    }