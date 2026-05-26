#!/bin/bash
set -e

VERSION=${1:-"1.0.0"}
OUTPUT=${2:-"linux-dashboard"}

echo "Building Linux Dashboard v${VERSION}..."

cd "$(dirname "$0")"

go build -ldflags=" \
  -s -w \
  -X main.version=${VERSION} \
" -o ${OUTPUT} ./cmd/ldm/

echo "✅ Done: ./${OUTPUT}"
echo "   Run: ./${OUTPUT}"
echo "   Open: http://127.0.0.1:19876"