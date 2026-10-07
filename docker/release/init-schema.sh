#!/bin/sh
set -eu

# PostgreSQL runs this only when initializing an empty data volume.
for migration in /migrations/*.up.sql; do
  psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    --set ON_ERROR_STOP=1 --single-transaction --file "$migration"
done
