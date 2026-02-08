#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

APP_NAME="${APP_NAME:-gofro-node}"
APP_USER="${APP_USER:-root}"
SERVICE_NAME="${SERVICE_NAME:-gofro-node}"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/.env}"
BIN_DIR="${BIN_DIR:-${ROOT_DIR}/bin}"
APP_BIN="${APP_BIN:-${BIN_DIR}/${APP_NAME}}"

XRAY_DIR="${XRAY_DIR:-${ROOT_DIR}/xray}"
XRAY_BIN="${XRAY_BIN:-${XRAY_DIR}/xray}"

require_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "error: run this script as root (sudo)."
    exit 1
  fi
}

detect_pkg_manager() {
  if command -v apt-get >/dev/null 2>&1; then
    PKG_INSTALL="apt-get update -y && apt-get install -y"
  elif command -v dnf >/dev/null 2>&1; then
    PKG_INSTALL="dnf install -y"
  elif command -v yum >/dev/null 2>&1; then
    PKG_INSTALL="yum install -y"
  elif command -v pacman >/dev/null 2>&1; then
    PKG_INSTALL="pacman -Syy --noconfirm"
  elif command -v apk >/dev/null 2>&1; then
    PKG_INSTALL="apk add --no-cache"
  else
    PKG_INSTALL=""
  fi
}

install_packages() {
  local packages=("$@")
  if [[ -z "${PKG_INSTALL}" ]]; then
    echo "warning: no supported package manager found; install dependencies manually: ${packages[*]}"
    return 0
  fi
  # shellcheck disable=SC2086
  ${PKG_INSTALL} "${packages[@]}"
}

ensure_go() {
  if command -v go >/dev/null 2>&1; then
    return 0
  fi
  detect_pkg_manager
  install_packages curl ca-certificates tar gzip
  install_packages golang || true

  if command -v go >/dev/null 2>&1; then
    return 0
  fi

  echo "warning: package manager didn't provide go; skipping go install."
  echo "warning: install go manually and rerun."
  exit 1
}

map_xray_arch() {
  case "$(uname -m)" in
    i386|i686) echo "32" ;;
    amd64|x86_64) echo "64" ;;
    armv5tel) echo "arm32-v5" ;;
    armv6l) echo "arm32-v6" ;;
    armv7|armv7l) echo "arm32-v7a" ;;
    aarch64|armv8) echo "arm64-v8a" ;;
    mips) echo "mips32" ;;
    mipsle) echo "mips32le" ;;
    mips64) echo "mips64" ;;
    mips64le) echo "mips64le" ;;
    ppc64) echo "ppc64" ;;
    ppc64le) echo "ppc64le" ;;
    riscv64) echo "riscv64" ;;
    s390x) echo "s390x" ;;
    *) echo "" ;;
  esac
}

latest_xray_tag() {
  curl -fsSL "https://api.github.com/repos/XTLS/Xray-core/releases/latest" \
    | sed -n 's/.*"tag_name":"\([^"]*\)".*/\1/p' | head -n 1
}

install_xray() {
  local arch tag url tmp
  arch="$(map_xray_arch)"
  if [[ -z "${arch}" ]]; then
    echo "error: unsupported architecture: $(uname -m)"
    exit 1
  fi

  install_packages curl unzip

  tag="$(latest_xray_tag)"
  if [[ -z "${tag}" ]]; then
    echo "error: failed to determine latest Xray version."
    exit 1
  fi

  url="https://github.com/XTLS/Xray-core/releases/download/${tag}/Xray-linux-${arch}.zip"
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/xray.zip" "${url}"
  unzip -q "${tmp}/xray.zip" -d "${tmp}"

  mkdir -p "${XRAY_DIR}"
  install -m 0755 "${tmp}/xray" "${XRAY_BIN}"
  if [[ -f "${tmp}/geoip.dat" ]]; then
    install -m 0644 "${tmp}/geoip.dat" "${XRAY_DIR}/geoip.dat"
  fi
  if [[ -f "${tmp}/geosite.dat" ]]; then
    install -m 0644 "${tmp}/geosite.dat" "${XRAY_DIR}/geosite.dat"
  fi

  rm -rf "${tmp}"
  echo "xray installed: ${XRAY_BIN}"
}

build_project() {
  ensure_go
  mkdir -p "${BIN_DIR}"
  (cd "${ROOT_DIR}" && CGO_ENABLED=0 go build -o "${APP_BIN}" ./cmd)
  echo "project built: ${APP_BIN}"
}

set_env_var() {
  local key="$1"
  local value="$2"
  if [[ -f "${ENV_FILE}" ]] && grep -qE "^${key}=" "${ENV_FILE}"; then
    sed -i.bak "s|^${key}=.*|${key}=${value}|" "${ENV_FILE}"
  else
    echo "${key}=${value}" >> "${ENV_FILE}"
  fi
}

create_env() {
  touch "${ENV_FILE}"
  local node_name="${NODE_NAME:-$(hostname)}"
  set_env_var "NODE_NAME" "${node_name}"
  set_env_var "XRAY_DEFAULT_CONFIG" "xconf/config.json"
  set_env_var "XRAY_CORE_PATH" "xray/xray"
  set_env_var "DEV_MODE" "false"
  echo ".env ready: ${ENV_FILE}"
}

create_service() {
  local unit="/etc/systemd/system/${SERVICE_NAME}.service"
  cat > "${unit}" <<EOF
[Unit]
Description=GoFroNet Node
After=network.target

[Service]
Type=simple
User=${APP_USER}
WorkingDirectory=${ROOT_DIR}
EnvironmentFile=${ENV_FILE}
ExecStart=${APP_BIN}
Restart=on-failure
RestartSec=3
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable "${SERVICE_NAME}"
  systemctl restart "${SERVICE_NAME}"
  echo "systemd service ready: ${SERVICE_NAME}"
}

main() {
  require_root
  detect_pkg_manager
  install_packages curl ca-certificates

  install_xray
  build_project
  create_env
  create_service

  echo "done."
}

main "$@"
