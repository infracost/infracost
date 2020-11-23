#!/bin/sh -le

tfjson=${1:-$tfjson}
tfplan=${2:-$tfplan}
use_tfstate=${3:-$use_tfstate}
tfdir=${4:-$tfdir}
tfflags=${5:-$tfflags}
percentage_threshold=${6:-$percentage_threshold}
pricing_api_endpoint=${7:-$pricing_api_endpoint}

infracost_cmd="infracost --no-color --log-level warn"
if [ ! -z "$tfjson" ]; then
  infracost_cmd="$infracost_cmd --tfjson $tfjson"
fi
if [ ! -z "$tfplan" ]; then
  infracost_cmd="$infracost_cmd --tfplan $tfplan"
fi
if [ "$use_tfstate" = "true" ] || [ "$use_tfstate" = "True" ] || [ "$use_tfstate" = "TRUE" ] ; then
  infracost_cmd="$infracost_cmd --use-tfstate"
fi
if [ ! -z "$tfdir" ]; then
  infracost_cmd="$infracost_cmd --tfdir $tfdir"
fi
if [ ! -z "$tfflags" ]; then
  infracost_cmd="$infracost_cmd --tfflags \"$tfflags\""
fi
if [ ! -z "$pricing_api_endpoint" ]; then
  infracost_cmd="$infracost_cmd --pricing-api-endpoint $pricing_api_endpoint"
fi
echo "$infracost_cmd" > infracost_cmd

echo "Running infracost on current branch using:"
echo "  $ $(cat infracost_cmd)"
current_branch_output=$(cat infracost_cmd | sh)
echo "$current_branch_output" > current_branch_infracost.txt
current_branch_monthly_cost=$(cat current_branch_infracost.txt | awk '/OVERALL TOTAL/ { print $NF }')
echo "::set-output name=current_branch_monthly_cost::$current_branch_monthly_cost"

echo "Switching to default branch"
git fetch --depth=1 origin master || git fetch --depth=1 origin main
git switch master || git switch main
git log -n1

echo "Running infracost on default branch using:"
echo "  $ $(cat infracost_cmd)"
default_branch_output=$(cat infracost_cmd | sh)
echo "$default_branch_output" > default_branch_infracost.txt
default_branch_monthly_cost=$(cat default_branch_infracost.txt | awk '/OVERALL TOTAL/ { print $NF }')
echo "::set-output name=default_branch_monthly_cost::$default_branch_monthly_cost"

percent_diff=$(echo "scale=4; $current_branch_monthly_cost / $default_branch_monthly_cost * 100 - 100" | bc)
absolute_percent_diff=$(echo $percent_diff | tr -d -)

if [ $(echo "$absolute_percent_diff > $percentage_threshold" | bc -l) == 1 ]; then
  change_word="increase"
  if [ $(echo "$percent_diff < 0" | bc -l) == 1 ]; then
    change_word="decrease"
  fi

  jq -Mnc --arg change_word $change_word \
          --arg absolute_percent_diff $(printf '%.1f\n' $absolute_percent_diff) \
          --arg default_branch_monthly_cost $default_branch_monthly_cost \
          --arg current_branch_monthly_cost $current_branch_monthly_cost \
          --arg diff "$(git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | tail -n +3)" \
          '{body: "Monthly cost estimate will \($change_word) by \($absolute_percent_diff)% (default branch $\($default_branch_monthly_cost) vs current branch $\($current_branch_monthly_cost))\n<details><summary>infracost diff</summary>\n\n```diff\n\($diff)\n```\n</details>\n"}' > diff_infracost.txt

  echo "Default branch and current branch diff ($absolute_percent_diff) is more than the percentage threshold ($percentage_threshold)"

  if [ ! -z "$GITHUB_ACTIONS" ]; then
    echo "Posting comment to GitHub"
    cat diff_infracost.txt | curl -sL -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/repos/$GITHUB_REPOSITORY/commits/$GITHUB_SHA/comments" > /dev/null
  elif [ ! -z "$GITLAB_CI" ]; then
    echo "Posting comment to GitLab"
    cat diff_infracost.txt | curl -sL -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$CI_SERVER_URL/api/v4/projects/$CI_PROJECT_ID/merge_requests/$CI_MERGE_REQUEST_IID/notes" > /dev/null
  elif [ ! -z "$CIRCLECI" ]; then
    echo "CircleCI integration: coming soon!"
  fi
else
  echo "Comment not posted as default branch and current branch diff ($absolute_percent_diff) is not more than the percentage threshold ($percentage_threshold)."
fi
