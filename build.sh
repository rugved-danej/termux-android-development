#!/bin/sh
# build.sh - Build standalone native ARM64 binary for tadev
# Author: Rugved Danej (@rugved-danej)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "[*] Building tadev standalone ARM64 binary..."
mkdir -p dist
rm -f dist/tadev
go build -ldflags="-s -w -X 'main.Version=1.0.0' -X 'main.Author=Rugved Danej'" -o dist/tadev main.go
chmod +x dist/tadev

echo "[+] Build complete: dist/tadev ($(ls -lh dist/tadev | awk '{print $5}'))"
