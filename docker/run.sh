#!/bin/bash

set -eu

if [ -f config.json ]; then
    exec ./gophish "$@"
fi

export ADMIN_LISTEN_URL="${ADMIN_LISTEN_URL:-0.0.0.0:3333}"
export ADMIN_USE_TLS="${ADMIN_USE_TLS:-true}"
export ADMIN_CERT_PATH="${ADMIN_CERT_PATH:-gophish_admin.crt}"
export ADMIN_KEY_PATH="${ADMIN_KEY_PATH:-gophish_admin.key}"
export PHISH_LISTEN_URL="${PHISH_LISTEN_URL:-0.0.0.0:80}"
export PHISH_USE_TLS="${PHISH_USE_TLS:-false}"
export PHISH_CERT_PATH="${PHISH_CERT_PATH:-example.crt}"
export PHISH_KEY_PATH="${PHISH_KEY_PATH:-example.key}"
export DB_NAME="${DB_NAME:-sqlite3}"
export DB_PATH="${DB_PATH:-${DB_FILE_PATH:-gophish.db}}"
export MIGRATIONS_PREFIX="${MIGRATIONS_PREFIX:-db/db_}"

exec ./gophish --config="" "$@"
