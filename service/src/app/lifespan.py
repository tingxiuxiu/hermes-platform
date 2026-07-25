from structlog import get_logger
from app.infrastructure.redis.session import redis_client, redis_pool

logger = get_logger(__name__)

# ===================== 4. 全局生命周期初始化/销毁 =====================
async def redis_init():
    """服务启动：初始化连接池并测试连通性"""
    await redis_client.ping()
    logger.info("Redis 连接池初始化完成，ping 正常")

async def redis_close():
    """服务关闭：销毁连接池，释放所有连接"""
    if redis_pool:
        await redis_pool.disconnect()
        logger.info("Redis 连接池已全部释放")
