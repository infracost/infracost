#!/usr/bin/env bash

COMMITISH=${1:-HEAD}
NO_DIRTY=${2:-false}
TAG=""
BUILD=""
DIRTY_SUFFIX=""

git update-index -q --refresh
if ! [ "$NO_DIRTY" = true ]  && ! git diff-files --quiet -- . ':!**/go.mod' ':!**/go.sum'; then
  DIRTY_SUFFIX="-dirty"
fi

if $(git rev-parse --is-shallow-repository); then
  git fetch --quiet --unshallow
fi

if git describe --tags --exact-match "${COMMITISH}" >/dev/null 2>&1; then
  TAG=$(git describe --tags --exact-match "${COMMITISH}")
elif git describe --tags "${COMMITISH}" > /dev/null 2>&1; then
  full_tag=$(git describe --tags "${COMMITISH}")
  commit_count=$(echo ${full_tag} | awk -F'-' '{print $2}')
  commit_sha=$(echo ${full_tag} | awk -F'-' '{print $3}')

  TAG=$(echo ${full_tag} | awk -F'-' '{print $1}')
  BUILD=$(echo +${commit_count}-${commit_sha})
fi


echo $TAG$BUILD$DIRTY_SUFFIX
