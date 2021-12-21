#!/usr/bin/env bash
set -e

# This script posts a comment on a pull request or commit.
# It first runs `infracost output` to generate the comment and then wraps
# compost (https://github.com/infracost/compost) to post the comment.
# 
# Usage:
#  ./comment.sh --platform <PLATFORM> --format <FORMAT> --behavior <BEHAVIOR> --target-type=<TARGET_TYPE> <INFRACOST_JSON_PATHS>
#
# Example:
#   ./comment.sh --platform gitlab --format gitlab-comment --behavior update --target-type=merge-request my_infracost_breakdown.json

# Parse the flags
parsed=$(getopt --options= --longoptions=platform:,format:,behavior:,target-type:,dry-run,skip-compost --name="$0" -- "$@")

eval set -- "$parsed"

declare -a input_paths
platform=
format=
behavior=upate
target_type=
dry_run=false
skip_compost=false

while true; do
  case "$1" in
    --platform ) platform="$2"; shift 2 ;;
    --format ) format="$2"; shift 2 ;;
    --behavior ) behavior="$2"; shift 2 ;;
    --target-type ) target_type="$2"; shift 2 ;;
    --dry-run ) dry_run=true; shift ;;
    --skip-compost ) skip_compost=true; shift ;;
    -- ) input_paths=("${@:2}"); break ;;
    * ) break ;;
  esac
done

export INFRACOST_CI_POST_CONDITION=${behavior}
body_file=$(mktemp)

# Handle multiple paths
input_path_flags=""
for input_path in "${input_paths[@]}"; do
  input_path_flags="$input_path_flags --path $input_path"
done

# shellcheck disable=SC2086
infracost output $input_path_flags --format $format --show-skipped --out-file $body_file

# Append a note to the comment about it being updated/replaced
if [ "$target_type" != "commit" ]; then
  if [ "$behavior" = "update" ]; then
    printf "\nThis comment will be updated when the cost estimate changes.\n\n" >> "$body_file"
  elif [ "$behavior" = "delete-and-new" ]; then
    printf "\nThis comment will be replaced when the cost estimate changes.\n\n" >> "$body_file"
  fi
fi

# Post the comment
flags=""
if [ -n "$platform" ]; then
  flags+=" --platform $platform"
fi
if [ -n "$target_type" ]; then
  flags+=" --target-type $target_type"
fi
if [ "$dry_run" = true ]; then
  flags+=" --dry-run"
fi

compost_cmd="compost autodetect $behavior --body-file $body_file$flags"

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

if [ "$skip_compost" = true ] || [ "$dry_run" = true ]; then
  echo "Comment output:"
  cat "$body_file"
fi
