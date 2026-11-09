#!/usr/bin/env bash
# Build semua file program dengan nomor versi dari file VERSION.
# Pemakaian: ./build.sh
set -euo pipefail
cd "$(dirname "$0")"

VER="$(tr -d '[:space:]' < VERSION)"
BUILD="$(date -u +%Y%m%d%H%M)"
LDFLAGS="-s -w -X main.version=${VER} -X main.buildTime=${BUILD}"

go vet ./...
echo "== tes otomatis"
go test -count=1 ./... >/dev/null || { echo "TES GAGAL: rilis dibatalkan. Jalankan: go test ./..."; exit 1; }
mkdir -p dist
for target in windows/amd64 windows/386 linux/amd64 linux/arm64 linux/arm; do
  os="${target%/*}"; arch="${target#*/}"; ext=""
  [ "$os" = windows ] && ext=".exe"
  echo "build ${os}/${arch}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" GOARM=7 go build -trimpath -ldflags "$LDFLAGS" \
    -o "dist/sch-server-${os}-${arch}${ext}" .
done
cp README.md dist/PANDUAN.md
cp CHANGELOG.md dist/CHANGELOG.md
echo "Selesai: SCH v${VER} (build ${BUILD}) di folder dist/"
