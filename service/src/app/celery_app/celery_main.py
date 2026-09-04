from celery import Celery

from app.config import settings

celery_app = Celery(
    "hermes-platform",
    broker=settings.CELERY_BROKER_URL or settings.redis_url,
    backend=settings.CELERY_RESULT_BACKEND or settings.redis_url,
)
celery_app.conf.update(
    task_always_eager=settings.CELERY_TASK_ALWAYS_EAGER,
    task_ignore_result=True,
    timezone="UTC",
    beat_schedule={
        "refresh-automation-dashboard-overview": {
            "task": "automation.dashboard.refresh_all",
            "schedule": settings.AUTOMATION_DASHBOARD_REFRESH_SECONDS,
        }
    },
)
