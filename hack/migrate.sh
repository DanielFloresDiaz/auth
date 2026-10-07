#!/usr/bin/env bash

DB_ENV=$1

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DATABASE="$DIR/database.yml"

export GOTRUE_DB_DRIVER="postgres"
export POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
export POSTGRES_PORT="${POSTGRES_PORT:-5437}"
export GOTRUE_DB_DATABASE_URL="postgres://auth_admin:root@${POSTGRES_HOST}:${POSTGRES_PORT}/$DB_ENV"
export GOTRUE_DB_MIGRATIONS_PATH=$DIR/../migrations

CONFIG="${AUTH_TEST_CONFIG:-$DIR/test.env}"
go run main.go migrate -c "$CONFIG"
