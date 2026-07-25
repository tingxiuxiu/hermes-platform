import json
from typing import Optional, Any
from app.infrastructure.redis.session import redis_client

async def set_json(
        key: str, data: Any, expire: Optional[int] = None
    ):
    """写入JSON序列化数据"""
    json_str = json.dumps(data, ensure_ascii=False)
    await redis_client.set(key, json_str, ex=expire)

async def get_json(key: str) -> Optional[Any]:
    """读取并自动反序列化JSON"""
    raw = await redis_client.get(key)
    if raw is None:
        return None
    return json.loads(raw)

async def delete(key: str):
    await redis_client.delete(key)

async def exists(key: str) -> bool:
    return await redis_client.exists(key) > 0
