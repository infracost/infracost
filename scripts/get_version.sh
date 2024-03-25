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

# Get the latest tag with a "v" prefix
LATEST_V_TAG=$(git tag --list 'v*' --sort=-v:refname | head -n 1)

if [ -n "$LATEST_V_TAG" ]; then
  if git describe --tags --exact-match "${COMMITISH}" 2>/dev/null | grep -q "^${LATEST_V_TAG}$"; then
    TAG=$LATEST_V_TAG
  elif git describe --tags "${COMMITISH}" --match 'v*' > /dev/null 2>&1; then
    full_tag=$(git describe --tags "${COMMITISH}" --match 'v*')
    commit_count=$(echo ${full_tag} | awk -F'-' '{print $(NF-1)}')
    commit_sha=$(echo ${full_tag} | awk -F'-' '{print $NF}')

    # Assuming the latest "v" prefixed tag is part of the description
    TAG=$(echo ${full_tag} | grep -o '^v[^-]*')
    BUILD=$(echo +${commit_count}-${commit_sha})
  fi
fi

echo $TAG$BUILD$DIRTY_SUFFIX
