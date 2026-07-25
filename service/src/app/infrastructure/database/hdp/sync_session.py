from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from app.config import settings

engine = create_engine(
    str(settings.sqlalchemy_uri),
    pool_size=settings.SQLALCHEMY_POOL_SIZE,
    max_overflow=settings.SQLALCHEMY_MAX_OVERFLOW,
    echo=settings.ENVIRONMENT != "production",
    pool_pre_ping=True,
    future=True,
)

SyncSessionLocal = sessionmaker(
    bind=engine,
    autocommit=False,
    autoflush=False,
)
