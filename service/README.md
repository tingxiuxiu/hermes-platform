#Hermes Platform - Data Plane Service

# Alembic

## check db

```
pdm run alembic current
```

## migration db

```
pdm run alembic revision --autogenerate -m "create account table"
```

## upgrade db

```
pdm run alembic upgrade head
```

## upgrade x version

```
pdm run alembic upgrade <version>
```

## downgrade x version

```
pdm run alembic downgrade <version>
```
