#!/usr/bin/env bash
# Local test environment without Docker: MariaDB, Redis stand-in, Kratos, imgproxy stand-in, API and front.
#
#   dev.sh setup   download MariaDB and Kratos, build the helper tools
#   dev.sh up      start everything (initializes and seeds the database on first run)
#   dev.sh down    stop everything
#   dev.sh status  show which services run
#   dev.sh seed    reload the test data
#   dev.sh reset   wipe the database and Kratos data, then start again
#   dev.sh logs    tail all logs
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE="$ROOT/.local-dev"
DATA="$STATE/data"
LOGS="$STATE/logs"
PIDS="$STATE/pids"

MARIADB_VERSION=11.8.9 # UUID_v4() needs 11.7+
KRATOS_VERSION=26.2.0

MARIADB="$STATE/mariadb/bin"
KRATOS="$STATE/kratos/kratos"
TOOLS="$STATE/bin/tools"

set -a
# shellcheck source=dev.env
source "$ROOT/dev/dev.env"
set +a

# go install puts golang-migrate here
PATH="$(go env GOBIN):$(go env GOPATH)/bin:$PATH"

mkdir -p "$DATA" "$LOGS" "$PIDS" "$STATE/cache" "$STATE/bin"

log() { echo "[dev] $*"; }

is_running() {
  local pidfile="$PIDS/$1.pid"
  [[ -f "$pidfile" ]] && kill -0 "$(cat "$pidfile")" 2>/dev/null
}

# start <name> <cwd> <command...>: detach the command in its own session
start() {
  local name="$1" cwd="$2"
  shift 2
  if is_running "$name"; then
    log "$name already running"
    return
  fi
  # the wrapper shell writes its own pid, then becomes the service, so the pid file is exact
  # redirect the whole subshell, or it keeps our stdout open and a piped caller never sees EOF
  (cd "$cwd" && exec setsid bash -c 'echo $$ >"$0"; exec "$@"' "$PIDS/$name.pid" "$@") \
    >"$LOGS/$name.log" 2>&1 </dev/null &
  sleep 0.2
}

stop() {
  local name="$1" pidfile="$PIDS/$1.pid"
  if is_running "$name"; then
    kill -- "-$(cat "$pidfile")" 2>/dev/null || kill "$(cat "$pidfile")" 2>/dev/null || true
    log "stopped $name"
  fi
  rm -f "$pidfile"
}

wait_for() {
  local name="$1" check="$2"
  for _ in $(seq 1 120); do
    if eval "$check" >/dev/null 2>&1; then
      log "$name ready"
      return
    fi
    sleep 1
  done
  log "$name did not become ready, see $LOGS/$name.log"
  exit 1
}

mysql_root() {
  "$MARIADB/mariadb" --skip-ssl --protocol=tcp -h127.0.0.1 -P"$DATABASE_PORT" -uroot "$@"
}

port_open() { (exec 3<>"/dev/tcp/127.0.0.1/$1") 2>/dev/null; }

cmd_setup() {
  if [[ ! -x "$MARIADB/mariadbd" ]]; then
    log "downloading MariaDB $MARIADB_VERSION"
    curl -fsSL -o "$STATE/cache/mariadb.tar.gz" \
      "https://archive.mariadb.org/mariadb-$MARIADB_VERSION/bintar-linux-systemd-x86_64/mariadb-$MARIADB_VERSION-linux-systemd-x86_64.tar.gz"
    mkdir -p "$STATE/mariadb"
    tar xzf "$STATE/cache/mariadb.tar.gz" -C "$STATE/mariadb" --strip-components=1
  fi
  if [[ ! -x "$KRATOS" ]]; then
    log "downloading Kratos $KRATOS_VERSION"
    curl -fsSL -o "$STATE/cache/kratos.tar.gz" \
      "https://github.com/ory/kratos/releases/download/v$KRATOS_VERSION/kratos_$KRATOS_VERSION-linux_sqlite_64bit.tar.gz"
    mkdir -p "$STATE/kratos"
    tar xzf "$STATE/cache/kratos.tar.gz" -C "$STATE/kratos" kratos
  fi
  if ! command -v migrate >/dev/null; then
    log "installing golang-migrate"
    go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  fi
  log "building helper tools"
  (cd "$ROOT/dev/tools" && go build -o "$TOOLS" .)
  if [[ ! -d "$ROOT/apps/front/node_modules" ]]; then
    log "installing front dependencies"
    (cd "$ROOT/apps/front" && bun install)
  fi
  log "setup done"
}

require_setup() {
  if [[ ! -x "$MARIADB/mariadbd" || ! -x "$KRATOS" || ! -x "$TOOLS" ]]; then
    log "run 'dev.sh setup' first"
    exit 1
  fi
}

up_mariadb() {
  local fresh=0
  if [[ ! -d "$DATA/mariadb" ]]; then
    log "initializing MariaDB data dir"
    "$STATE/mariadb/scripts/mariadb-install-db" --no-defaults --basedir="$STATE/mariadb" --datadir="$DATA/mariadb" \
      --auth-root-authentication-method=normal --skip-test-db >"$LOGS/mariadb-install.log" 2>&1
    fresh=1
  fi
  start mariadb "$STATE" "$MARIADB/mariadbd" --no-defaults --basedir="$STATE/mariadb" --datadir="$DATA/mariadb" \
    --bind-address=127.0.0.1 --port="$DATABASE_PORT" --socket="$DATA/mariadb.sock" \
    --pid-file="$DATA/mariadb.pid" --character-set-server=utf8mb4 --skip-name-resolve
  wait_for mariadb 'mysql_root -e "SELECT 1"'

  if [[ "$(mysql_root -N -e "SHOW DATABASES LIKE '$DATABASE_NAME'")" == "" ]]; then
    mysql_root -e "CREATE DATABASE \`$DATABASE_NAME\` CHARACTER SET utf8mb4"
    fresh=1
  fi
  log "running migrations"
  migrate -database "mysql://root:@tcp(127.0.0.1:$DATABASE_PORT)/$DATABASE_NAME" \
    -path "$ROOT/apps/api/db/migration" up
  if [[ "$fresh" == 1 ]]; then
    NEED_SEED=1
  fi
}

up_kratos() {
  sed "s#@ROOT@#$ROOT#g" "$ROOT/dev/kratos/kratos.yml" >"$STATE/kratos.yml"
  start kratos "$STATE" env "DSN=sqlite://$DATA/kratos.sqlite?_fk=true&mode=rwc" \
    sh -c "'$KRATOS' migrate sql -e --yes -c '$STATE/kratos.yml' && exec '$KRATOS' serve -c '$STATE/kratos.yml' --dev"
  wait_for kratos 'curl -sf -m 5 http://localhost:4434/health/ready'
}

# Prints the id of the test identity, creating the identity when it does not exist.
ensure_test_identity() {
  local id
  id="$(curl -sf "http://localhost:4434/admin/identities?credentials_identifier=$DEV_USER_EMAIL" |
    python3 -c 'import json,sys; d=json.load(sys.stdin); print(d[0]["id"] if d else "")')"
  if [[ -z "$id" ]]; then
    id="$(curl -sf -X POST http://localhost:4434/admin/identities -H 'Content-Type: application/json' -d "{
      \"schema_id\": \"default\",
      \"traits\": {\"email\": \"$DEV_USER_EMAIL\", \"username\": \"tester\"},
      \"credentials\": {\"password\": {\"config\": {\"password\": \"$DEV_USER_PASSWORD\"}}}
    }" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')"
  fi
  echo "$id"
}

cmd_seed() {
  local owner_id
  owner_id="$(ensure_test_identity)"
  log "seeding test data (test user $DEV_USER_EMAIL, id $owner_id)"
  { echo "SET @owner_id = '$owner_id';"; cat "$ROOT/dev/seed.sql"; } | mysql_root "$DATABASE_NAME"
}

cmd_up() {
  require_setup
  NEED_SEED=0
  up_mariadb
  start redis "$ROOT" "$TOOLS" redis 127.0.0.1:6379
  start imgproxy "$ROOT" "$TOOLS" imgproxy 127.0.0.1:8100
  up_kratos
  if [[ "$NEED_SEED" == 1 ]]; then
    cmd_seed
  fi

  start api "$ROOT/apps/api" go run ./cmd
  wait_for api 'curl -sf -m 5 http://localhost:8080/v1/health'
  start front "$ROOT/apps/front" env PORT=3000 bun run dev
  wait_for front 'curl -s -m 60 -o /dev/null http://localhost:3000/'

  cat <<MSG

Local environment is up.
  front    http://localhost:3000
  api      http://localhost:8080/v1  (swagger: http://localhost:8080/swagger/index.html)
  kratos   http://localhost:4433
  login    $DEV_USER_EMAIL / $DEV_USER_PASSWORD
  logs     $LOGS
MSG
}

cmd_down() {
  for name in front api kratos imgproxy redis mariadb; do
    stop "$name"
  done
}

cmd_status() {
  for name in mariadb redis imgproxy kratos api front; do
    if is_running "$name"; then echo "up    $name"; else echo "down  $name"; fi
  done
}

cmd_reset() {
  cmd_down
  rm -rf "$DATA"
  cmd_up
}

case "${1:-}" in
  setup) cmd_setup ;;
  up) cmd_up ;;
  down) cmd_down ;;
  status) cmd_status ;;
  seed) cmd_seed ;;
  reset) cmd_reset ;;
  logs) tail -n 20 -F "$LOGS"/*.log ;;
  *) sed -n '2,10p' "${BASH_SOURCE[0]}"; exit 1 ;;
esac
