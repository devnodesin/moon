
#!/bin/bash
# Moon Build Script (Docker-based)
set -e

if [ "$EUID" -ne 0 ]; then
    echo "Insufficient privileges."
    echo ""
    echo "Try running this script with sudo:"
    echo "  sudo ./samples/build.sh"
    echo ""
    exit 1
fi

# Check for Docker
if ! command -v docker >/dev/null 2>&1; then
    echo "[ERROR] Docker is not installed or not in PATH."
    exit 1
fi

echo "[INFO] Building Moon binary using Docker..."

docker run --rm \
  -v "$(pwd):/app" \
  -v "$(pwd)/.gocache:/gocache" \
  -w /app \
  -e GOCACHE=/gocache \
  golang:latest sh -c "go build -buildvcs=false -o moon ./cmd/moon"

