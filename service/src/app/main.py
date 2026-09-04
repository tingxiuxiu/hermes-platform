from fastapi import FastAPI

from app.config import settings
from app.observability import init_observability
from app.auth.auth_router import auth_router
from app.automation.router import automation_router

app = FastAPI(
    title=settings.PROJECT_NAME,
    docs_url=f"{settings.API_V1_STR}/docs",
    redoc_url=f"{settings.API_V1_STR}/redoc",
)

# init_observability(app)

app.include_router(auth_router, prefix=settings.API_V1_STR)
app.include_router(automation_router, prefix=settings.API_V1_STR)
