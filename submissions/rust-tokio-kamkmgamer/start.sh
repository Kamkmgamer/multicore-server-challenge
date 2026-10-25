#!/usr/bin/env bash
# Runs the server in the foreground. Config: DATABASE_URL, JWT_SECRET, HOST, PORT.
# Optional process tuning: DB_CONNS, DB_WRITE_CONNS, APP_THREADS,
# BCRYPT_THREADS, BCRYPT_BATCH_MS, WRITE_BATCH_US, READ_BATCH_SIZE, READ_BATCH_US,
# READ_SQL_BATCH, READ_ROUTE_SLACK, READ_TAG_ROUTE, READ_COMMENT_SET, READ_TAG_BATCH.
# No additional configuration is required.
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
# Let the Rust runtime derive APP_THREADS, DB_CONNS, DB_WRITE_CONNS, and
# BCRYPT_THREADS from the available cgroup CPUs. On the fixed eight-core box
# this selects the proven 2/6/2/1 profile; local hosts can use their capacity.
export BCRYPT_BATCH_MS="${BCRYPT_BATCH_MS:-160}"
export WRITE_BATCH_US="${WRITE_BATCH_US:-1000}"
export READ_BATCH_SIZE="${READ_BATCH_SIZE:-32}"
export READ_BATCH_US="${READ_BATCH_US:-0}"
export READ_SQL_BATCH="${READ_SQL_BATCH:-1}"
export READ_ROUTE_SLACK="${READ_ROUTE_SLACK:-8}"
export READ_TAG_ROUTE="${READ_TAG_ROUTE:-1}"
export READ_COMMENT_SET="${READ_COMMENT_SET:-1}"
export READ_TAG_BATCH="${READ_TAG_BATCH:-1}"
exec "${CARGO_TARGET_DIR:-$DIR/target}/release/conduit"
