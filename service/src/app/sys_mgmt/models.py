from app.models import Base
from sqlalchemy import (
    Integer,
    SmallInteger,
    String,
    Text,
    UniqueConstraint,
    Index,
)
from sqlalchemy.orm import mapped_column


class SysDict(Base):
    __tablename__ = "sys_dict"

    id = mapped_column(Integer, primary_key=True, autoincrement=True)
    dict_type = mapped_column(String(64), nullable=False, unique=True)
    code = mapped_column(SmallInteger, nullable=False, comment="字典编码")
    label = mapped_column(String(128), nullable=False, comment="前端标签")
    dict_name = mapped_column(String(128), nullable=False)
    description = mapped_column(Text, nullable=True)
    category = mapped_column(String(64), nullable=True, comment="字典分类")
    sort_order = mapped_column(Integer, nullable=False, default=0, comment="排序")
    color = mapped_column(String(32), nullable=True, comment="前端颜色")
    status = mapped_column(String(32), nullable=False, default="active")
    __table_args__ = (
        UniqueConstraint("dict_type", "code", name="uq_sys_dict_type_code"),
        Index("idx_sys_dict_type", "dict_type"),
    )
