#!/bin/sh -le

tfjson=${1:-$tfjson}
tfplan=${2:-$tfplan}
use_tfstate=${3:-$use_tfstate}
tfdir=${4:-$tfdir}
tfflags=${5:-$tfflags}
percentage_threshold=${6:-$percentage_threshold}
pricing_api_endpoint=${7:-$pricing_api_endpoint}

INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}

infracost_cmd="infracost --no-color"
if [ ! -z "$tfjson" ]; then
  infracost_cmd="$infracost_cmd --tfjson $tfjson"
fi
if [ ! -z "$tfplan" ]; then
  infracost_cmd="$infracost_cmd --tfplan $tfplan"
fi
if [ "$use_tfstate" = "true" ] || [ "$use_tfstate" = "True" ] || [ "$use_tfstate" = "TRUE" ]; then
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
current_branch_monthly_cost=$(cat current_branch_infracost.txt | awk '/OVERALL TOTAL/ { printf("%.2f",$NF) }')
echo "::set-output name=current_branch_monthly_cost::$current_branch_monthly_cost"

current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" = "master" ] || [ "$current_branch" = "main" ]; then
  echo "Exiting as the current branch was the default branch so nothing more to do."
  exit 0
fi
echo "Switching to default branch"
git fetch --depth=1 origin master &>/dev/null || git fetch --depth=1 origin main &>/dev/null || echo "Could not fetch default branch from origin, no problems, switching to it..."
git switch master &>/dev/null || git switch main &>/dev/null
git log -n1

echo "Running infracost on default branch using:"
echo "  $ $(cat infracost_cmd)"
default_branch_output=$(cat infracost_cmd | sh)
echo "$default_branch_output" > default_branch_infracost.txt
default_branch_monthly_cost=$(cat default_branch_infracost.txt | awk '/OVERALL TOTAL/ { printf("%.2f",$NF) }')
echo "::set-output name=default_branch_monthly_cost::$default_branch_monthly_cost"

percent_diff=$(echo "scale=4; $current_branch_monthly_cost / $default_branch_monthly_cost * 100 - 100" | bc)
absolute_percent_diff=$(echo $percent_diff | tr -d -)

if [ $(echo "$absolute_percent_diff > $percentage_threshold" | bc -l) = 1 ]; then
  change_word="increase"
  if [ $(echo "$percent_diff < 0" | bc -l) = 1 ]; then
    change_word="decrease"
  fi

  jq -Mnc --arg change_word $change_word \
          --arg absolute_percent_diff $(printf '%.1f\n' $absolute_percent_diff) \
          --arg default_branch_monthly_cost $default_branch_monthly_cost \
          --arg current_branch_monthly_cost $current_branch_monthly_cost \
          --arg diff "$(git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | tail -n +3)" \
          '{body: "Monthly cost estimate will \($change_word) by \($absolute_percent_diff)% (default branch $\($default_branch_monthly_cost) vs current branch $\($current_branch_monthly_cost))\n<details><summary>infracost diff</summary>\n\n```diff\n\($diff)\n```\n</details>\n"}' > diff_infracost.txt

  echo "Default branch and current branch diff ($absolute_percent_diff) is more than the percentage threshold ($percentage_threshold)."

  GITHUB_API_URL=${GITHUB_API_URL:-https://api.github.com}
  BITBUCKET_API_URL=${BITBUCKET_API_URL:-https://api.bitbucket.org}

  if [ ! -z "$GITHUB_ACTIONS" ]; then
    if [ "$GITHUB_EVENT_NAME" == "pull_request" ]; then
      GITHUB_SHA=$(cat $GITHUB_EVENT_PATH | jq -r .pull_request.head.sha)
    fi  
    echo "Posting comment to GitHub commit $GITHUB_SHA"
    cat diff_infracost.txt | curl -L -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "Authorization: token $GITHUB_TOKEN" \
        "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/commits/$GITHUB_SHA/comments"

  elif [ ! -z "$GITLAB_CI" ]; then
    echo "Posting comment to GitLab"
    cat diff_infracost.txt | curl -L -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$CI_SERVER_URL/api/v4/projects/$CI_PROJECT_ID/merge_requests/$CI_MERGE_REQUEST_IID/notes"

  elif [ ! -z "$CIRCLECI" ]; then
    if echo $CIRCLE_REPOSITORY_URL | grep -Eiq github; then
      echo "Posting comment from CircleCI to GitHub commit $CIRCLE_SHA1"
      cat diff_infracost.txt | curl -L -X POST -d @- \
          -H "Content-Type: application/json" \
          -H "Authorization: token $GITHUB_TOKEN" \
          "$GITHUB_API_URL/repos/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/commits/$CIRCLE_SHA1/comments"

    elif echo $CIRCLE_REPOSITORY_URL | grep -Eiq bitbucket; then
      echo "Posting comment from CircleCI to BitBucket commit $CIRCLE_SHA1"
      # BitBucket comments require a different JSON format and don't support HTML 
      jq -Mnc --arg change_word $change_word \
              --arg absolute_percent_diff $(printf '%.1f\n' $absolute_percent_diff) \
              --arg default_branch_monthly_cost $default_branch_monthly_cost \
              --arg current_branch_monthly_cost $current_branch_monthly_cost \
              --arg diff "$(git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | tail -n +3)" \
              '{content: {raw: "Monthly cost estimate will \($change_word) by \($absolute_percent_diff)% (default branch $\($default_branch_monthly_cost) vs current branch $\($current_branch_monthly_cost))\n\n```diff\n\($diff)\n```\n"}}' > diff_infracost.txt

      # BITBUCKET_TOKEN must be set to "myusername:my_app_password"
      cat diff_infracost.txt | curl -L -X POST -d @- \
          -H "Content-Type: application/json" \
          -u $BITBUCKET_TOKEN \
          "$BITBUCKET_API_URL/2.0/repositories/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/commit/$CIRCLE_SHA1/comments"
    fi
  fi
else
  echo "Comment not posted as default branch and current branch diff ($absolute_percent_diff) is not more than the percentage threshold ($percentage_threshold)."
fi
