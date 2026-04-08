#!/usr/bin/env bash

set -Eeuo pipefail

APP_NAME="${APP_NAME:-gobkd}"
SERVICE_NAME="${SERVICE_NAME:-$APP_NAME}"
INSTALL_ROOT="${INSTALL_ROOT:-/opt/$APP_NAME}"
REMOVE_SERVICE_FILE="${REMOVE_SERVICE_FILE:-0}"
CONFIRM_UNINSTALL="${CONFIRM_UNINSTALL:-}"

log() {
  printf '[uninstall] %s\n' "$*"
}

fail() {
  printf '[uninstall] error: %s\n' "$*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Linux" ]]; then
  fail "this script must run on Linux"
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

main() {
  local service_file

  if [[ "$CONFIRM_UNINSTALL" != "YES" ]]; then
    cat >&2 <<EOF
[uninstall] this will remove:
[uninstall] - service: $SERVICE_NAME
[uninstall] - install root: $INSTALL_ROOT
[uninstall] rerun with CONFIRM_UNINSTALL=YES to continue
EOF
    exit 1
  fi

  if service_exists; then
    log "stopping systemd service: $SERVICE_NAME"
    run_root systemctl stop "$SERVICE_NAME" || true
    run_root systemctl disable "$SERVICE_NAME" || true
  else
    log "systemd service not found: $SERVICE_NAME"
  fi

  if [[ "$REMOVE_SERVICE_FILE" == "1" ]]; then
    service_file="/etc/systemd/system/$SERVICE_NAME.service"
    if [[ -f "$service_file" ]]; then
      log "removing service file: $service_file"
      run_root rm -f "$service_file"
    fi
  fi

  if command -v systemctl >/dev/null 2>&1; then
    run_root systemctl daemon-reload || true
  fi

  if [[ -e "$INSTALL_ROOT" ]]; then
    log "removing install root: $INSTALL_ROOT"
    run_root rm -rf "$INSTALL_ROOT"
  else
    log "install root not found: $INSTALL_ROOT"
  fi

  log "uninstall complete"
}

main "$@"
