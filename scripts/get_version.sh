#!/usr/bin/env bash

COMMITISH=${1:-HEAD}
SUFFIX=""

git update-index -q --refresh
if ! git diff-files --quiet -- . ':!**/go.mod' ':!**/go.sum'; then
  SUFFIX="-dirty"
fi

git fetch --tags
if git describe --tags --exact-match "${COMMITISH}" >/dev/null 2>&1; then
  TAG=$(git describe --tags --exact-match "${COMMITISH}")
elif git describe --tags "${COMMITISH}" > /dev/null 2>&1; then
  TAG=$(git describe --tags "${COMMITISH}")
fi

echo $TAG$SUFFIX
