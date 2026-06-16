#!/bin/sh
# code-runner コンテナの起動スクリプト。
# 同居の使い捨て PostgreSQL（SQL 演習用 / socket 専用 / 非 superuser student）を起動してから
# coderunner 本体を exec する。PostgreSQL は build 時に initdb 済み（/pgdata）。
set -e

PGDATA="${PGDATA:-/pgdata}"
SOCKDIR="${CODE_PG_HOST:-/tmp/pgsock}"

# /tmp が runtime で空（tmpfs マウント等）でも socket dir を必ず用意する。
mkdir -p "$SOCKDIR"

# socket 専用（-h '' で TCP を開かない）で起動。ネットワークからは到達不能。
pg_ctl -D "$PGDATA" -o "-k $SOCKDIR -h ''" -w -t 30 start

exec /coderunner
