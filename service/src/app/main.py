from fastapi import FastAPI

from app.config import settings
from app.observability import init_observability
from app.api.v1.router import api_router

app = FastAPI(
    title=settings.PROJECT_NAME,
    docs_url=f"{settings.API_V1_STR}/docs",
    redoc_url=f"{settings.API_V1_STR}/redoc"
)

# init_observability(app)

app.include_router(api_router, prefix=settings.API_V1_STR)
