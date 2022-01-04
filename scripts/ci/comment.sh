#!/usr/bin/env bash
set -e

# This script posts a comment on a pull request or commit.
# It first runs `infracost output` to generate the comment and then wraps
# compost (https://github.com/infracost/compost) to post the comment.
#
# The script uses the following environment variables:
#   - COMMENT_FORMAT: The format of the comment as supported by `infracost output`.
#       Default: github-comment
#       Options: github-comment, gitlab-comment, azure-repos-comment
#   - COMMENT_BEHAVIOR: The behavior of the comment as supported by `compost`.
#       Default: update
#       Options:
#         update: create a single comment and update it. The "quietest" option.
#         delete-and-new: delete previous comments and create a new one.
#         hide-and-new: minimize previous comments and create a new one (only supported by GitHub).
#         new: create a new cost estimate comment on every push.
#   - COMMENT_TARGET_TYPE: Which objects should be commented on
#       Default: (empty) - will try and find a pull/merge request, if not it will comment on a commit
#       Options: pull-request, merge-request, commit.
#   - COMMENT_TAG: Customize the comment tag.
#       This is added to the comment as a markdown comment (hidden) to detect
#       the previously posted comments. This is useful if you have multiple
#       workflows that post comments to the same pull request or commit.
#   - COMMENT_PLATFORM: Only required if we need to limit the compost auto-detection to a specific platform.
#       By default this will be auto-detected.
#
# For testing:
#   - COMMENT_DRY_RUN: Run compost in dry run mode so comments aren't posted, updated or deleted.
#   - COMMENT_SKIP_COMPOST: Skip the call to compost
# 
# Usage:
#   COMMENT_FORMAT=<FORMAT> COMMENT_BEHAVIOR=<BEHAVIOR> COMMENT_TARGET_TYPE=<TARGET-TYPE> ./comment.sh <INFRACOST_JSON_PATHS>
#
# Example:
#   COMMENT_FORMAT=gitlab-comment COMMENT_BEHAVIOR=update COMMENT_TARGET_TYPE=merge-request ./comment.sh infracost-dev.json infracost-prod.json

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
