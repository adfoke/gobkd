#!/usr/bin/env bash

set -Eeuo pipefail

APP_NAME="${APP_NAME:-gobkd}"
SERVICE_NAME="${SERVICE_NAME:-$APP_NAME}"
BIN_NAME="${BIN_NAME:-$APP_NAME}"
REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
INSTALL_ROOT="${INSTALL_ROOT:-/opt/$APP_NAME}"
RELEASES_DIR="${RELEASES_DIR:-$INSTALL_ROOT/releases}"
SHARED_DIR="${SHARED_DIR:-$INSTALL_ROOT/shared}"
CURRENT_LINK="${CURRENT_LINK:-$INSTALL_ROOT/current}"
BUILD_DIR="${BUILD_DIR:-$REPO_ROOT/.build}"
ENV_SOURCE="${ENV_SOURCE:-$REPO_ROOT/.env}"
ENV_TARGET="${ENV_TARGET:-$SHARED_DIR/.env}"
DATA_DIR="${DATA_DIR:-$SHARED_DIR/data}"
HEALTHCHECK_URL="${HEALTHCHECK_URL:-http://127.0.0.1:8080/healthz}"
GOOS_VALUE="${GOOS:-linux}"
GOARCH_VALUE="${GOARCH:-$(go env GOARCH)}"
RESTART_SERVICE="${RESTART_SERVICE:-auto}"
HEALTHCHECK_RETRIES="${HEALTHCHECK_RETRIES:-20}"
HEALTHCHECK_INTERVAL="${HEALTHCHECK_INTERVAL:-1}"

log() {
  printf '[deploy] %s\n' "$*"
}

fail() {
  printf '[deploy] error: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"
}

if [[ "$(uname -s)" != "Linux" ]]; then
  fail "this script must run on Linux"
fi

need_cmd go
need_cmd install
need_cmd ln
need_cmd mv
need_cmd date

if command -v gcc >/dev/null 2>&1; then
  CC_BIN="gcc"
elif command -v cc >/dev/null 2>&1; then
  CC_BIN="cc"
else
  fail "missing C compiler: gcc or cc is required for sqlite3 cgo build"
fi

if command -v sudo >/dev/null 2>&1 && [[ "$(id -u)" -ne 0 ]]; then
  SUDO=(sudo)
else
  SUDO=()
fi

run_root() {
  "${SUDO[@]}" "$@"
}

service_exists() {
  command -v systemctl >/dev/null 2>&1 && systemctl cat "$SERVICE_NAME" >/dev/null 2>&1
}

wait_for_healthcheck() {
  need_cmd curl

  local attempt=1
  while (( attempt <= HEALTHCHECK_RETRIES )); do
    if curl --fail --silent --show-error "$HEALTHCHECK_URL" >/dev/null; then
      log "healthcheck passed: $HEALTHCHECK_URL"
      return 0
    fi

    sleep "$HEALTHCHECK_INTERVAL"
    attempt=$((attempt + 1))
  done

  fail "healthcheck failed: $HEALTHCHECK_URL"
}

main() {
  local release_id release_dir tmp_release tmp_link

  release_id="$(date +%Y%m%d%H%M%S)"
  release_dir="$RELEASES_DIR/$release_id"

  mkdir -p "$BUILD_DIR"
  tmp_release="$(mktemp -d "$BUILD_DIR/release.XXXXXX")"
  trap 'rm -rf "$tmp_release"' EXIT

  log "building $APP_NAME for $GOOS_VALUE/$GOARCH_VALUE"
  (
    cd "$REPO_ROOT"
    CGO_ENABLED=1 CC="$CC_BIN" GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
      go build -trimpath -ldflags="-s -w" -o "$tmp_release/$BIN_NAME" ./cmd/server
  )

  run_root install -d -m 755 "$INSTALL_ROOT" "$RELEASES_DIR" "$SHARED_DIR" "$DATA_DIR" "$release_dir"
  run_root install -m 755 "$tmp_release/$BIN_NAME" "$release_dir/$BIN_NAME"

  if [[ ! -f "$ENV_TARGET" ]]; then
    if [[ -f "$ENV_SOURCE" ]]; then
      log "initializing shared .env from $ENV_SOURCE"
      run_root install -m 600 "$ENV_SOURCE" "$ENV_TARGET"
    elif [[ -f "$REPO_ROOT/.env.example" ]]; then
      fail "missing $ENV_SOURCE, copy .env.example to .env and fill required values first"
    else
      fail "missing $ENV_SOURCE"
    fi
  fi

  run_root ln -sfn "$ENV_TARGET" "$release_dir/.env"
  run_root ln -sfn "$DATA_DIR" "$release_dir/data"

  tmp_link="$INSTALL_ROOT/.current.tmp"
  run_root ln -sfn "$release_dir" "$tmp_link"
  run_root mv -Tf "$tmp_link" "$CURRENT_LINK"

  log "release deployed: $release_dir"
  log "current -> $release_dir"

  if [[ "$RESTART_SERVICE" == "never" ]]; then
    log "service restart skipped by RESTART_SERVICE=never"
    return 0
  fi

  if service_exists; then
    log "restarting systemd service: $SERVICE_NAME"
    run_root systemctl daemon-reload
    run_root systemctl restart "$SERVICE_NAME"
    wait_for_healthcheck
    return 0
  fi

  log "systemd service not found: $SERVICE_NAME"
  log "create /etc/systemd/system/$SERVICE_NAME.service from deploy/gobkd.service.example, then rerun"
}

main "$@"
