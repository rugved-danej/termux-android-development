#!/bin/sh
# build.sh - Build standalone native ARM64 binary for tadev
# Author: Rugved Danej (@rugved-danej)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

printf "\033[35m•\033[0m Building tadev standalone ARM64 binary...\n"
mkdir -p dist
rm -f dist/tadev
go build -ldflags="-s -w -X 'main.Version=1.0.0' -X 'main.Author=Rugved Danej'" -o dist/tadev main.go
chmod +x dist/tadev

printf "\033[32m✓\033[0m Build complete: dist/tadev (%s)\n" "$(ls -lh dist/tadev | awk '{print $5}')"
