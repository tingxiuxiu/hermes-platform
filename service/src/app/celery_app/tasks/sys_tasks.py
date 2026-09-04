import json

from app.models.hdp.sys_dict import SysDict
from app.celery_app.celery_main import celery_app
from app.infrastructure.redis.sync_session import sync_redis_client as redis_client
from app.infrastructure.database.hdp.sync_session import SyncSessionLocal


def reflush_sys_dict_cache():
    """
    刷新系统字典缓存
    """
    # 查询所有的系统字典数据
    with SyncSessionLocal() as session:
        sys_dicts = session.query(SysDict).all()
        # 将字典数据存入缓存
        for sys_dict in sys_dicts:
            cache_key = f"sys_dict:{sys_dict.dict_type}:{sys_dict.code}"
            cache_value = {
                "label": sys_dict.label,
                "dict_name": sys_dict.dict_name,
                "description": sys_dict.description,
                "category": sys_dict.category,
                "sort_order": sys_dict.sort_order,
                "color": sys_dict.color,
                "status": sys_dict.status,
            }
            redis_client.set(cache_key, json.dumps(cache_value, ensure_ascii=False))
        # 按category存入列表，方便后端给前端列表
        for category in set(d.category for d in sys_dicts if d.category):
            category_dicts = [
                {
                    "dict_type": d.dict_type,
                    "code": d.code,
                    "label": d.label,
                    "dict_name": d.dict_name,
                    "description": d.description,
                    "sort_order": d.sort_order,
                    "color": d.color,
                    "status": d.status,
                }
                for d in sys_dicts
                if d.category == category
            ]
            redis_client.set(
                f"sys_dict:category:{category}",
                json.dumps(category_dicts, ensure_ascii=False),
            )
