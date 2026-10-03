#!/bin/sh
# Выкатить domashka на сервер: make deploy SERVER=user@host
#
# Образ собирается здесь, под архитектуру сервера, и уезжает по ssh; на сервере нужен только
# Docker с compose (исходники, git, make, Node и Go там не нужны). Рядом с образом в
# ~/domashka/ кладётся compose.yaml, контейнер поднимается и проверяется (healthy).
#
#   PORT=8081 make deploy SERVER=…   — другой порт на сервере (тогда и в nginx.conf)
#   ALLOW_DIRTY=1 make deploy …      — выкатить с незакоммиченными правками
set -eu

server=${1:?укажи сервер: make deploy SERVER=user@host}
dir=${DEPLOY_DIR:-domashka}
port=${PORT:-8080}

if [ -z "${ALLOW_DIRTY:-}" ] && [ -n "$(git status --porcelain)" ]; then
	echo "есть незакоммиченные правки — выкатывается то, что в git (или ALLOW_DIRTY=1)" >&2
	exit 1
fi
tag=$(git rev-parse --short HEAD)

case $(ssh "$server" uname -m) in
x86_64 | amd64) platform=linux/amd64 ;;
aarch64 | arm64) platform=linux/arm64 ;;
*) echo "неизвестная архитектура сервера" >&2; exit 1 ;;
esac

echo "→ собираю domashka:$tag под $platform"
docker buildx build --platform "$platform" -t "domashka:$tag" -t domashka:latest --load .

echo "→ передаю образ на $server"
docker save "domashka:$tag" | gzip | ssh "$server" 'gunzip | docker load'

echo "→ запускаю"
ssh "$server" "mkdir -p '$dir'"
scp -q compose.yaml "$server:$dir/compose.yaml"
# Прежние образы остаются с тегами по коммитам — откат: docker tag domashka:<коммит> domashka:latest
ssh "$server" "docker tag 'domashka:$tag' domashka:latest && cd '$dir' && PORT='$port' docker compose up -d --no-build --wait"

echo "✓ domashka:$tag работает на $server (127.0.0.1:$port, путь /domashka/)"
