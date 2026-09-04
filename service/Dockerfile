FROM python:3.14-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PDM_CHECK_UPDATE=false \
    PDM_HOME=/root/.local/share/pdm \
    PATH=/root/.local/bin:$PATH

RUN pip install --no-cache-dir pdm

WORKDIR /app

COPY service/pyproject.toml ./pyproject.toml
COPY service/README.md ./README.md
COPY service/alembic.ini ./alembic.ini
COPY service/alembic ./alembic
COPY service/src ./src

RUN pdm install --prod --no-editable

EXPOSE 8080