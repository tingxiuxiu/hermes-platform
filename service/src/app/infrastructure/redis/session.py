from redis.asyncio import Redis, ConnectionPool
from app.config import settings

# ===================== 1. 全局连接池单例 =====================
# 全局唯一连接池，所有Redis客户端共享一套连接池
redis_pool = ConnectionPool(
        host=settings.REDIS_HOST,
        port=settings.REDIS_PORT,
        db=settings.REDIS_DB,
        password=settings.REDIS_PASSWORD or None,
        socket_timeout=settings.REDIS_SOCKET_TIMEOUT,
        retry_on_timeout=settings.REDIS_RETRY_ON_TIMEOUT,
        max_connections=settings.REDIS_MAX_CONNECTIONS,
        decode_responses=False,  # 统一二进制，序列化自己处理
    )

# ===================== 2. 获取Redis客户端（复用全局池） =====================
redis_client = Redis(connection_pool=redis_pool)
