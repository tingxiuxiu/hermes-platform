from datetime import datetime

from sqlalchemy.orm import DeclarativeBase

from typing import Optional, List, Dict, Any
from sqlalchemy import (
    BigInteger, String, Text, SmallInteger, Boolean, 
    TIMESTAMP, ForeignKey, CheckConstraint, UniqueConstraint, Index, Table, Column
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func


class Base(DeclarativeBase):
    created_at: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True), 
        server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True), 
        server_default=func.now(), 
        onupdate=func.now()
    )
