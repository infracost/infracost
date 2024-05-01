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
  else
    full_tag=$(git describe --tags "${COMMITISH}" --match 'v*')
    if [[ $full_tag =~ ^(v[0-9]+\.[0-9]+\.[0-9]+)-([0-9]+)-g([0-9a-f]+)$ ]]; then
      TAG=${BASH_REMATCH[1]}
      BUILD=$(echo +${BASH_REMATCH[2]}-${BASH_REMATCH[3]})
    else
      TAG=${full_tag}
    fi
  fi
fi

echo $TAG$BUILD$DIRTY_SUFFIX
