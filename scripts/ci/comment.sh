#!/usr/bin/env bash
set -e

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# This script is DEPRECATED and is no longer maintained.
#
# Please follow our guide: https://www.infracost.io/docs/guides/gitlab_ci_migration/ to migrate
# to our new GitHub Actions integration: https://gitlab.com/infracost/infracost-gitlab-ci/
#
# This script will be removed September 2022.
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
echo "Warning: this script is deprecated and will be removed in Sep 2022."
echo "Please visit https://www.infracost.io/docs/guides/gitlab_ci_migration/ for instructions on how to upgrade."
echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

format=${COMMENT_FORMAT:-github-comment}
behavior=${COMMENT_BEHAVIOR:-update}
target_type=$COMMENT_TARGET_TYPE
tag=$COMMENT_TAG
platform=$COMMENT_PLATFORM

# Used for testing
dry_run=${COMMENT_DRY_RUN:-false}
skip_compost=${COMMENT_SKIP_COMPOST:-false}

# Build flags for Infracost to handle multiple paths
if [ "$#" -eq 0 ]; then
  echo "Expecting at least one Infracost JSON file as argument"
  exit 1
fi

input_path_flags=""
for input_path in "$@"; do
  input_path_flags="$input_path_flags --path $input_path"
done

# Run infracost output to generate the comment
export INFRACOST_CI_POST_CONDITION=$behavior
body_file=$(mktemp)
# shellcheck disable=SC2086
infracost_cmd="infracost output $input_path_flags --format $format --show-skipped --out-file $body_file"
echo "Running infracost output:" >&2
echo "  $infracost_cmd" >&2
eval "$infracost_cmd"

# Append a note to the comment about it being updated/replaced
if [ "$target_type" != "commit" ]; then
  if [ "$behavior" = "update" ]; then
    printf "\nThis comment will be updated when the cost estimate changes.\n\n" >> "$body_file"
  elif [ "$behavior" = "delete-and-new" ]; then
    printf "\nThis comment will be replaced when the cost estimate changes.\n\n" >> "$body_file"
  fi
fi

# Generate the compost flags
flags=""
if [ -n "$platform" ]; then
  flags+=" --platform $platform"
fi
if [ -n "$target_type" ]; then
  flags+=" --target-type $target_type"
fi
if [ -n "$tag" ]; then
  flags+=" --tag $tag"
fi
if [ "$dry_run" = true ]; then
  flags+=" --dry-run"
fi

# Build the compost command
compost_cmd="compost autodetect $behavior --body-file $body_file$flags"

# Run compost
# shellcheck disable=SC2086
if [ "$skip_compost" = true ]; then
  echo "Skipping compost:" >&2
  echo "  $compost_cmd" >&2
else
  if [ "$dry_run" = true ]; then
    echo "Running compost in dry run mode:" >&2
  else
    echo "Running compost:" >&2
  fi
  echo "  $compost_cmd" >&2
  eval "$compost_cmd"
fi

# If testing then output the comment
if [ "$skip_compost" = true ] || [ "$dry_run" = true ]; then
  echo "Comment output:"
  cat "$body_file"
fi
