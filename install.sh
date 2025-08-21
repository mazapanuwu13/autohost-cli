#!/usr/bin/env bash
set -euo pipefail

REPO="mazapanuwu13/autohost-cli"
BIN_NAME="autohost"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"

# Detectar OS y ARCH
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"   # linux, darwin
ARCH="$(uname -m)"                              # x86_64, aarch64, armv7l...

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "❌ Arquitectura no soportada: $ARCH"; exit 1 ;;
esac

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

# Última versión
LATEST_TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"

if [ -z "${LATEST_TAG:-}" ]; then
  echo "❌ No se pudo obtener la última versión. ¿Hay releases?"
  exit 1
fi

echo "📦 Instalando $BIN_NAME $LATEST_TAG para $OS-$ARCH..."

ASSET="${BIN_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET}"

# Descargar binario
curl -fsSL -o "${TMP_DIR}/${BIN_NAME}" "${URL}"

chmod +x "${TMP_DIR}/${BIN_NAME}"

echo "🚚 Moviendo a ${BIN_DIR}..."
if [ -w "${BIN_DIR}" ]; then
  mv "${TMP_DIR}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
else
  sudo mv "${TMP_DIR}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
fi

echo "✅ Instalación completa: $(command -v ${BIN_NAME})"
echo "👉 Ejecuta: ${BIN_NAME} --help"
