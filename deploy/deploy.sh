#!/bin/sh
set -eu

server=${1:?usage: make deploy SERVER=user@host}
dir=${DEPLOY_DIR:-domashka}
port=${PORT:-8080}

if [ -z "${ALLOW_DIRTY:-}" ] && [ -n "$(git status --porcelain)" ]; then
	echo "there are uncommitted changes; only committed code is deployed (or set ALLOW_DIRTY=1)" >&2
	exit 1
fi
tag=$(git rev-parse --short HEAD)

case $(ssh "$server" uname -m) in
x86_64 | amd64) platform=linux/amd64 ;;
aarch64 | arm64) platform=linux/arm64 ;;
*) echo "unsupported server architecture" >&2; exit 1 ;;
esac

echo "→ building domashka:$tag for $platform"
docker buildx build --platform "$platform" -t "domashka:$tag" -t domashka:latest --load .

echo "→ sending the image to $server"
docker save "domashka:$tag" | gzip | ssh "$server" 'gunzip | docker load'

echo "→ starting"
ssh "$server" "mkdir -p '$dir'"
scp -q compose.yaml "$server:$dir/compose.yaml"
ssh "$server" "docker tag 'domashka:$tag' domashka:latest && cd '$dir' && PORT='$port' docker compose up -d --no-build --wait"

echo "✓ domashka:$tag is running on $server (127.0.0.1:$port, path /domashka/)"
