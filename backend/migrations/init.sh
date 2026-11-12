#!/bin/bash
set -e

echo "Running PostgreSQL initialization scripts"

# Подключаемся к PostgreSQL и выполняем миграции
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f /docker-entrypoint-initdb.d/001_initial_schema.sql

echo "PostgreSQL initialization completed" 