#!/bin/sh -le

# This scripts runs infracost on the current branch then the master branch. It uses `git diff`
# to post a pull-request comment showing the cost estimate difference whenever a percentage
# threshold is crossed.
# Usage docs: https://www.infracost.io/docs/integrations/
# It supports: GitHub Actions, GitLab CI, CircleCI with GitHub and Bitbucket, Bitbucket Pipelines
# For Bitbucket: BITBUCKET_TOKEN must be set to "myusername:my_app_password", the password needs to have Read scope
#   on "Repositories" and "Pull Requests" so it can post comments. Using a Bitbucket App password
#   (https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) is recommended.

# Set variables based on the order for GitHub Actions, or the env value for other CIs
terraform_json_file=${1:-$terraform_json_file}
terraform_plan_file=${2:-$terraform_plan_file}
terraform_use_state=${3:-$terraform_use_state}
terraform_dir=${4:-$terraform_dir}
terraform_plan_flags=${5:-$terraform_plan_flags}
percentage_threshold=${6:-$percentage_threshold}
pricing_api_endpoint=${7:-$pricing_api_endpoint}
usage_file=${8:-$usage_file}

# Set defaults
percentage_threshold=${percentage_threshold:-0}
GITHUB_API_URL=${GITHUB_API_URL:-https://api.github.com}
BITBUCKET_API_URL=${BITBUCKET_API_URL:-https://api.bitbucket.org}
# Export as it's used by infracost, not this script
export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
export INFRACOST_CI_DIFF=true

if [ ! -z "$GIT_SSH_KEY" ]; then
  echo "Setting up private Git SSH key so terraform can access your private modules."
  mkdir -p .ssh
  echo "${GIT_SSH_KEY}" > .ssh/git_ssh_key
  chmod 600 .ssh/git_ssh_key
  export GIT_SSH_COMMAND="ssh -i $(pwd)/.ssh/git_ssh_key -o 'StrictHostKeyChecking=no'"
fi

# Bitbucket Pipelines don't have a unique env so use this to detect it
if [ ! -z "$BITBUCKET_BUILD_NUMBER" ]; then
  BITBUCKET_PIPELINES=true
fi

post_bitbucket_comment () {
  # Bitbucket comments require a different JSON format and don't support HTML
  jq -Mnc --arg change_word $change_word \
          --arg absolute_percent_diff $(printf '%.1f\n' $absolute_percent_diff) \
          --arg default_branch_monthly_cost $default_branch_monthly_cost \
          --arg current_branch_monthly_cost $current_branch_monthly_cost \
          --arg diff "$(git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | sed 1,2d | sed 3,5d)" \
          '{content: {raw: "Monthly cost estimate will \($change_word) by \($absolute_percent_diff)% (default branch $\($default_branch_monthly_cost) vs current branch $\($current_branch_monthly_cost))\n\n```diff\n\($diff)\n```\n"}}' > diff_infracost.txt

  cat diff_infracost.txt | curl -L -X POST -d @- \
            -H "Content-Type: application/json" \
            -u $BITBUCKET_TOKEN \
            "$BITBUCKET_API_URL/2.0/repositories/$1"
}

infracost_cmd="infracost --no-color"
if [ ! -z "$terraform_json_file" ]; then
  echo "WARNING: we do not recommend using terraform_json_file as it doesn't work with this diff script, use terraform_dir instead."
  infracost_cmd="$infracost_cmd --terraform-json-file $terraform_json_file"
fi
if [ ! -z "$terraform_plan_file" ]; then
  echo "WARNING: we do not recommend using terraform_plan_file as it doesn't work with this diff script, use terraform_dir instead."
  infracost_cmd="$infracost_cmd --terraform-plan-file $terraform_plan_file"
fi
if [ "$terraform_use_state" = "true" ] || [ "$terraform_use_state" = "True" ] || [ "$terraform_use_state" = "TRUE" ]; then
  echo "WARNING: we do not recommend using terraform_use_state as it doesn't work with this diff script, use terraform_dir without this instead."
  infracost_cmd="$infracost_cmd --terraform-use-state"
fi
if [ ! -z "$terraform_dir" ]; then
  infracost_cmd="$infracost_cmd --terraform-dir $terraform_dir"
fi
if [ ! -z "$terraform_plan_flags" ]; then
  infracost_cmd="$infracost_cmd --terraform-plan-flags \"$terraform_plan_flags\""
fi
if [ ! -z "$pricing_api_endpoint" ]; then
  infracost_cmd="$infracost_cmd --pricing-api-endpoint $pricing_api_endpoint"
fi
if [ ! -z "$usage_file" ]; then
  infracost_cmd="$infracost_cmd --usage-file $usage_file"
fi
echo "$infracost_cmd" > infracost_cmd

echo "Running infracost on current branch using:"
echo "  $ $(cat infracost_cmd)"
current_branch_output=$(cat infracost_cmd | sh)
# The sed is needed to cause the header line to be different between current_branch_infracost and
# default_branch_infracost, otherwise git diff removes it as its an identical line
echo "$current_branch_output" | sed 's/MONTHLY COST/MONTHLY COST /' > current_branch_infracost.txt
current_branch_monthly_cost=$(cat current_branch_infracost.txt | awk '/OVERALL TOTAL/ { gsub(",",""); printf("%.2f",$NF) }')
echo "::set-output name=current_branch_monthly_cost::$current_branch_monthly_cost"

current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" = "master" ] || [ "$current_branch" = "main" ]; then
  echo "Exiting as the current branch was the default branch so nothing more to do."
  exit 0
fi
if [ ! -z "$BITBUCKET_PIPELINES" ]; then
  echo "Configuring git remote for Bitbucket Pipelines"
  git config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"
fi
echo "Switching to default branch"
git fetch --depth=1 origin master &>/dev/null || git fetch --depth=1 origin main &>/dev/null || echo "Could not fetch default branch from origin, no problems, switching to it..."
git switch master &>/dev/null || git switch main &>/dev/null || (echo "Error: could not switch to branch master or main" && exit 1)
git log -n1

echo "Running infracost on default branch using:"
echo "  $ $(cat infracost_cmd)"
default_branch_output=$(cat infracost_cmd | sh)
echo "$default_branch_output" > default_branch_infracost.txt
default_branch_monthly_cost=$(cat default_branch_infracost.txt | awk '/OVERALL TOTAL/ { gsub(",",""); printf("%.2f",$NF) }')
echo "::set-output name=default_branch_monthly_cost::$default_branch_monthly_cost"

if [ $(echo "$default_branch_monthly_cost > 0" | bc -l) = 1 ]; then
  percent_diff=$(echo "scale=4; $current_branch_monthly_cost / $default_branch_monthly_cost * 100 - 100" | bc)
else
  echo "Default branch has no cost, setting percent_diff=100 to force a comment"
  percent_diff=100
fi
absolute_percent_diff=$(echo $percent_diff | tr -d -)

if [ $(echo "$absolute_percent_diff > $percentage_threshold" | bc -l) = 1 ]; then
  change_word="increase"
  if [ $(echo "$percent_diff < 0" | bc -l) = 1 ]; then
    change_word="decrease"
  fi

  comment_key=body
  if [ ! -z "$GITLAB_CI" ]; then
    comment_key=note
  fi
  jq -Mnc --arg comment_key $comment_key \
          --arg change_word $change_word \
          --arg absolute_percent_diff $(printf '%.1f\n' $absolute_percent_diff) \
          --arg default_branch_monthly_cost $default_branch_monthly_cost \
          --arg current_branch_monthly_cost $current_branch_monthly_cost \
          --arg diff "$(git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | sed 1,2d | sed 3,5d)" \
          '{($comment_key): "Monthly cost estimate will \($change_word) by \($absolute_percent_diff)% (default branch $\($default_branch_monthly_cost) vs current branch $\($current_branch_monthly_cost))\n<details><summary>infracost diff</summary>\n\n```diff\n\($diff)\n```\n</details>\n"}' > diff_infracost.txt

  echo "Default branch and current branch diff ($absolute_percent_diff) is more than the percentage threshold ($percentage_threshold)."

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
    echo "Posting comment to GitLab commit $CI_COMMIT_SHA"
    cat diff_infracost.txt | curl -L -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$CI_SERVER_URL/api/v4/projects/$CI_PROJECT_ID/repository/commits/$CI_COMMIT_SHA/comments"
        # Previously we posted to the merge request, using the comment_key=body above:
        # "$CI_SERVER_URL/api/v4/projects/$CI_PROJECT_ID/merge_requests/$CI_MERGE_REQUEST_IID/notes"

  elif [ ! -z "$CIRCLECI" ]; then
    if echo $CIRCLE_REPOSITORY_URL | grep -Eiq github; then
      echo "Posting comment from CircleCI to GitHub commit $CIRCLE_SHA1"
      cat diff_infracost.txt | curl -L -X POST -d @- \
          -H "Content-Type: application/json" \
          -H "Authorization: token $GITHUB_TOKEN" \
          "$GITHUB_API_URL/repos/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/commits/$CIRCLE_SHA1/comments"

    elif echo $CIRCLE_REPOSITORY_URL | grep -Eiq bitbucket; then
      if [ ! -z "$CIRCLE_PULL_REQUEST" ]; then
        BITBUCKET_PR_ID=$(echo $CIRCLE_PULL_REQUEST | sed 's/.*pull-requests\///')
        echo "Posting comment from CircleCI to Bitbucket pull-request $BITBUCKET_PR_ID"
        post_bitbucket_comment "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/pullrequests/$BITBUCKET_PR_ID/comments"

      else
        echo "Posting comment from CircleCI to Bitbucket commit $CIRCLE_SHA1"
        post_bitbucket_comment "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/commit/$CIRCLE_SHA1/comments"
      fi

    else
      echo "Error: CircleCI is not being used with GitHub or Bitbucket!"
    fi

  elif [ ! -z "$BITBUCKET_PIPELINES" ]; then
    if [ ! -z "$BITBUCKET_PR_ID" ]; then
      echo "Posting comment to Bitbucket pull-request $BITBUCKET_PR_ID"
      post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/pullrequests/$BITBUCKET_PR_ID/comments"

    else
      echo "Posting comment to Bitbucket commit $BITBUCKET_COMMIT"
      post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/commit/$BITBUCKET_COMMIT/comments"
    fi
  fi
else
  echo "Comment not posted as default branch and current branch diff ($absolute_percent_diff) is not more than the percentage threshold ($percentage_threshold)."
fi
