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

echo "creating no_ff_merge repo..."
curl -u 'test:test' -XPOST -H 'Content-Type: application/json' -d '{"name":"no_ff_merge"}' http://localhost:3000/api/v1/user/repos

echo "populating no_ff_merge repo..."
tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'tmpdir')
cd $tmpdir
export GIT_COMMITTER_NAME=test
export GIT_COMMITTER_EMAIL=test@test.com
export GIT_AUTHOR_NAME=test
export GIT_AUTHOR_EMAIL=test@test.com
git init --initial-branch=master
git commit -m "feat: initial commit" --allow-empty
git tag v1.0.0
git switch -C feature
sleep 1
git commit -m "feat: feature" --allow-empty
git switch master
git merge --no-ff feature --no-edit
git push http://test:test@localhost:3000/test/no_ff_merge.git master --tags
cd -
rm -rf $tmpdir
