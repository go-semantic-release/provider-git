#!/bin/bash

set -euo pipefail

docker stop gitea || true
docker rm gitea || true

git checkout ./test/gitea/gitea.db

docker run -d --name gitea -p 3000:3000 \
  -e USER_UID=1000 -e USER_GID=1000 \
  -v $(pwd)/test/gitea/conf/:/data/gitea/conf/ \
  -v $(pwd)/test/gitea/gitea.db:/data/gitea/gitea.db \
  gitea/gitea:1.15.6

sleep 10

echo "creating test repo..."
curl -u 'test:test' -XPOST -H 'Content-Type: application/json' -d '{"name":"test"}' http://localhost:3000/api/v1/user/repos
