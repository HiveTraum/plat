#!/bin/sh
set -e

until pg_isready -h "$POSTGRES_HOST" -U "$POSTGRES_USER"; do
  sleep 1
done

psql "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}/${POSTGRES_DB}" \
  -tc "SELECT 1 FROM pg_database WHERE datname = 'kratos'" | grep -q 1 \
  || psql "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}/${POSTGRES_DB}" \
    -c "CREATE DATABASE kratos"
