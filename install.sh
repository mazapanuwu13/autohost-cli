#!/usr/bin/env bash
set -euo pipefail

REPO="mazapanuwu13/autohost-cli"
BIN_NAME="autohost"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
VERSION="${VERSION:-}"   # opcional: export VERSION=v0.1.0 para fijar una versión/tag

# Detectar OS/ARCH
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"   # linux, darwin
ARCH_RAW="$(uname -m)"                          # x86_64, aarch64, etc.
case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "❌ Arquitectura no soportada: $ARCH_RAW"; exit 1 ;;
esac

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

fetch_release_bin() {
  local tag="$1"
  local asset="${BIN_NAME}-${OS}-${ARCH}"
  local url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
  echo "⬇️  Descargando binario de release: $url"
  curl -fLsS -o "${TMP_DIR}/${BIN_NAME}" "$url"
}

install_from_release() {
  # Si VERSION viene seteada, úsala; si no, intenta latest
  if [ -n "${VERSION}" ]; then
    fetch_release_bin "${VERSION}"
  else
    echo "🔎 Buscando última versión (releases)..."
    LATEST_TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
      | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)"
    if [ -z "${LATEST_TAG:-}" ]; then
      echo "ℹ️  No hay releases publicados."
      return 1
    fi
    fetch_release_bin "${LATEST_TAG}"
  fi

  chmod +x "${TMP_DIR}/${BIN_NAME}"
  echo "🚚 Instalando en ${BIN_DIR}..."
  mkdir -p "${BIN_DIR}"
  if [ -w "${BIN_DIR}" ]; then
    mv "${TMP_DIR}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
  else
    sudo mv "${TMP_DIR}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
  fi
  echo "✅ Instalación completa: $(command -v ${BIN_NAME})"
}

install_from_source() {
  echo "🛠  Compilando desde código (go install)..."
  if ! command -v go >/dev/null 2>&1; then
    echo "❌ Necesitas Go instalado para esta ruta (sudo apt-get install -y golang)."
    exit 1
  fi
  # Usa el módulo del repo (requiere go.mod con module github.com/mazapanuwu13/autohost-cli)
  local mod="github.com/${REPO}"
  local target="${mod}/cmd/${BIN_NAME}"

  if [ -n "${VERSION}" ]; then
    GO111MODULE=on go install "${target}@${VERSION}"
  else
    GO111MODULE=on go install "${target}@latest"
  fi

  # GOPATH/bin o GOBIN
  BIN_SRC="$(go env GOBIN || true)"
  if [ -z "$BIN_SRC" ]; then
    BIN_SRC="$(go env GOPATH)/bin"
  fi

  if [ ! -f "${BIN_SRC}/${BIN_NAME}" ]; then
    echo "❌ No se encontró ${BIN_NAME} en ${BIN_SRC}. ¿compiló bien?"
    exit 1
  fi

  echo "🚚 Moviendo a ${BIN_DIR}..."
  mkdir -p "${BIN_DIR}"
  if [ -w "${BIN_DIR}" ]; then
    mv "${BIN_SRC}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
  else
    sudo mv "${BIN_SRC}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
  fi

  echo "✅ Instalación completa: $(command -v ${BIN_NAME})"
}

# intento con release; si falla, compilo desde fuente
if ! install_from_release; then
  install_from_source
fi

echo "👉 Ejecuta: ${BIN_NAME} --help"
