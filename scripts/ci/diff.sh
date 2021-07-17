#!/bin/sh -le

# This script is used in infracost CI/CD integrations. It posts pull-request comments showing cost estimate diffs.
# Usage docs: https://www.infracost.io/docs/integrations/cicd
# It supports: GitHub Actions, GitLab CI, CircleCI with GitHub and Bitbucket, Bitbucket Pipelines, Azure DevOps with TfsGit repos and GitHub
# For Bitbucket: BITBUCKET_TOKEN must be set to "myusername:my_app_password", the password needs to have Read scope
#   on "Repositories" and "Pull Requests" so it can post comments. Using a Bitbucket App password
#   (https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) is recommended.

process_args () {
  # Set variables based on the order for GitHub Actions, or the env value for other CIs
  path=${1:-$path}
  terraform_plan_flags=${2:-$terraform_plan_flags}
  terraform_workspace=${3:-$terraform_workspace}
  usage_file=${4:-$usage_file}
  config_file=${5:-$config_file}
  percentage_threshold=${6:-$percentage_threshold}
  post_condition=${7:-$post_condition}
  show_skipped=${8:-$show_skipped}

  # Validate post_condition
  if ! echo "$post_condition" | jq empty; then
    echo "Error: post_condition contains invalid JSON"
  fi

  # Set defaults
  if [ ! -z "$percentage_threshold" ] && [ ! -z "$post_condition" ]; then
    echo "Warning: percentage_threshold is deprecated, using post_condition instead"
  elif [ ! -z "$percentage_threshold" ]; then
    post_condition="{\"percentage_threshold\": $percentage_threshold}"
    echo "Warning: percentage_threshold is deprecated and will be removed in v0.9.0, please use post_condition='{\"percentage_threshold\": \"0\"}'"
  else
    post_condition=${post_condition:-'{"has_diff": true}'}
  fi
  if [ ! -z "$post_condition" ] && [ "$(echo "$post_condition" | jq '.percentage_threshold')" != "null" ]; then
    percentage_threshold=$(echo "$post_condition" | jq -r '.percentage_threshold')
  fi
  percentage_threshold=${percentage_threshold:-0}
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}
  GITHUB_API_URL=${GITHUB_API_URL:-https://api.github.com}

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
}

build_breakdown_cmd () {
  breakdown_cmd="${INFRACOST_BINARY} breakdown --no-color --format json"

  if [ ! -z "$path" ]; then
    breakdown_cmd="$breakdown_cmd --path $path"
  fi
  if [ ! -z "$terraform_plan_flags" ]; then
    breakdown_cmd="$breakdown_cmd --terraform-plan-flags \"$terraform_plan_flags\""
  fi
  if [ ! -z "$terraform_workspace" ]; then
    breakdown_cmd="$breakdown_cmd --terraform-workspace $terraform_workspace"
  fi
  if [ ! -z "$usage_file" ]; then
    breakdown_cmd="$breakdown_cmd --usage-file $usage_file"
  fi
  if [ ! -z "$config_file" ]; then
    breakdown_cmd="$breakdown_cmd --config-file $config_file"
  fi
  echo "$breakdown_cmd"
}

build_output_cmd () {
  output_cmd="${INFRACOST_BINARY} output --no-color --format diff --path $1"
  if [ ! -z "$show_skipped" ]; then
    # The "=" is important as otherwise the value of the flag is ignored by the CLI
    output_cmd="$output_cmd --show-skipped=$show_skipped"
  fi
  echo "${output_cmd}"
}

format_cost () {
  cost=$1
    
  if [ -z "$cost" ] || [ "${cost}" = "null" ]; then
    echo "-"
  elif [ "$(echo "$cost < 100" | bc -l)" = 1 ]; then
    printf "$%0.2f" "$cost"
  else
    printf "$%0.0f" "$cost"
  fi
}

build_msg () {
  include_html=$1
  
  change_word="increase"
  change_sym="+"
    change_emoji="📈"
  if [ $(echo "$total_monthly_cost < ${past_total_monthly_cost}" | bc -l) = 1 ]; then
    change_word="decrease"
    change_sym=""
    change_emoji="📉"
  fi
  
  percent_display=""
  if [ ! -z "$percent" ]; then
    percent_display="$(printf "%.0f" $percent)"
    percent_display=" (${change_sym}${percent_display}%%)"
  fi
  
  msg="💰 Infracost estimate: **monthly cost will ${change_word} by $(format_cost $diff_cost)$percent_display** ${change_emoji}\n"
  msg="${msg}\n"
  msg="${msg}Previous monthly cost: $(format_cost $past_total_monthly_cost)\n"
  msg="${msg}New monthly cost: $(format_cost $total_monthly_cost)\n"
  msg="${msg}\n"
  
  if [ "$include_html" = true ]; then
    msg="${msg}<details>\n"
    msg="${msg}  <summary><strong>Infracost output</strong></summary>\n"
  else
    msg="${msg}**Infracost output:**\n"
  fi
    
  msg="${msg}\n"
  msg="${msg}\`\`\`\n"
  msg="${msg}$(echo "${diff_output}" | sed "s/%/%%/g")\n"
  msg="${msg}\`\`\`\n"
  
  if [ "$include_html" = true ]; then
    msg="${msg}</details>\n"
  fi
  
  printf "$msg"
}

post_to_github () {
  if [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
    GITHUB_SHA=$(cat $GITHUB_EVENT_PATH | jq -r .pull_request.head.sha)
  fi

  if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN is required to post comment to GitHub"
  else
    echo "Posting comment to GitHub commit $GITHUB_SHA"
    msg="$(build_msg true)"
    jq -Mnc --arg msg "$msg" '{"body": "\($msg)"}' | curl -L -X POST -d @- \
      -H "Content-Type: application/json" \
      -H "Authorization: token $GITHUB_TOKEN" \
      "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/commits/$GITHUB_SHA/comments"
  fi
}

post_to_gitlab () {
  echo "Posting comment to GitLab commit $CI_COMMIT_SHA"
  msg="$(build_msg true)"
  jq -Mnc --arg msg "$msg" '{"note": "\($msg)"}' | curl -L -X POST -d @- \
    -H "Content-Type: application/json" \
    -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
    "$CI_SERVER_URL/api/v4/projects/$CI_PROJECT_ID/repository/commits/$CI_COMMIT_SHA/comments"
}

post_bitbucket_comment () {
  msg="$(build_msg false)"
  jq -Mnc --arg msg "$msg" '{"content": {"raw": "\($msg)"}}' | curl -L -X POST -d @- \
    -H "Content-Type: application/json" \
    -u $BITBUCKET_TOKEN \
    "https://api.bitbucket.org/2.0/repositories/$1"
}

post_to_circle_ci () {
  if echo $CIRCLE_REPOSITORY_URL | grep -Eiq github; then
    echo "Posting comment from CircleCI to GitHub commit $CIRCLE_SHA1"
    msg="$(build_msg true)"
    jq -Mnc --arg msg "$msg" '{"body": "\($msg)"}' | curl -L -X POST -d @- \
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
}

post_to_bitbucket () {
  if [ ! -z "$BITBUCKET_PR_ID" ]; then
    echo "Posting comment to Bitbucket pull-request $BITBUCKET_PR_ID"
    post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/pullrequests/$BITBUCKET_PR_ID/comments"
  else
    echo "Posting comment to Bitbucket commit $BITBUCKET_COMMIT"
    post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/commit/$BITBUCKET_COMMIT/comments"
  fi
}

post_to_azure_devops () {
  if [ "$BUILD_REASON" = "PullRequest" ]; then
    if [ "$BUILD_REPOSITORY_PROVIDER" = "GitHub" ]; then
      echo "Posting comment to Azure DevOps GitHub pull-request $SYSTEM_PULLREQUEST_PULLREQUESTNUMBER"
      GITHUB_REPOSITORY=$BUILD_REPOSITORY_NAME
      GITHUB_SHA=$SYSTEM_PULLREQUEST_SOURCECOMMITID
      post_to_github
    elif [ "$BUILD_REPOSITORY_PROVIDER" = "TfsGit" ]; then
      echo "Posting comment to Azure DevOps repo pull-request $SYSTEM_PULLREQUEST_PULLREQUESTID"
      msg="$(build_msg true)"
      jq -Mnc --arg msg "$msg" '{"comments": [{"parentCommentId": 0, "content": "\($msg)", "commentType": 1}], "status": 4}' | curl -L -X POST -d @- \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $SYSTEM_ACCESSTOKEN" \
        "$SYSTEM_COLLECTIONURI$SYSTEM_TEAMPROJECT/_apis/git/repositories/$BUILD_REPOSITORY_ID/pullRequests/$SYSTEM_PULLREQUEST_PULLREQUESTID/threads?api-version=6.0"
    else 
      echo "Posting comments to Azure DevOps $BUILD_REPOSITORY_PROVIDER is not supported, email hello@infracost.io for help"
    fi
  else
    echo "Posting comment to Azure DevOps $BUILD_REASON is not supported, email hello@infracost.io for help"
  fi
}

post_to_slack () {
  echo "Posting comment to Slack"
  msg="$(build_msg false)"
  jq -Mnc --arg msg "$msg" '{"text": "\($msg)"}' | curl -L -X POST -d @- \
    -H "Content-Type: application/json" \
    "$SLACK_WEBHOOK_URL"
}

load_github_env () {
  export VCS_REPO_URL=$GITHUB_SERVER_URL/$GITHUB_REPOSITORY
  
  github_event=$(cat $GITHUB_EVENT_PATH)

  if [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
    GITHUB_SHA=$(echo $github_event | jq -r .pull_request.head.sha)
    export VCS_PULL_REQUEST_URL=$(echo $github_event | jq -r .pull_request.html_url)
  else
    export VCS_PULL_REQUEST_URL=$(curl -s \
      -H "Accept: application/vnd.github.groot-preview+json" \
      -H "Authorization: token $GITHUB_TOKEN" \
      $GITHUB_API_URL/repos/$GITHUB_REPOSITORY/commits/$GITHUB_SHA/pulls \
      | jq -r '. | map(select(.state == "open")) | . |= sort_by(.updated_at) | reverse | .[0].html_url')
  fi
}

load_gitlab_env () {
  export VCS_REPO_URL=$CI_REPOSITORY_URL
  
  first_mr=$(echo $CI_OPEN_MERGE_REQUESTS | cut -d',' -f1)
  repo=$(echo $first_mr | cut -d'!' -f1)
  mr_number=$(echo $first_mr | cut -d'!' -f2)
  export VCS_PULL_REQUEST_URL=$CI_SERVER_URL/$repo/merge_requests/$mr_number
}

load_circle_ci_env () {
  export VCS_REPO_URL=$CIRCLE_REPOSITORY_URL
}

load_azure_devops_env () {
  export VCS_REPO_URL=$BUILD_REPOSITORY_URI
}

cleanup () {
  rm -f infracost_breakdown.json infracost_breakdown_cmd infracost_output_cmd
}

# MAIN

process_args "$@"

# Load env variables
if [ ! -z "$GITHUB_ACTIONS" ]; then
  load_github_env
elif [ ! -z "$GITLAB_CI" ]; then
  load_gitlab_env
elif [ ! -z "$CIRCLECI" ]; then
  load_circle_ci_env
elif [ ! -z "$SYSTEM_COLLECTIONURI" ]; then
  load_azure_devops_env
fi

infracost_breakdown_cmd=$(build_breakdown_cmd)
echo "$infracost_breakdown_cmd" > infracost_breakdown_cmd

echo "Running infracost breakdown using:"
echo "  $ $(cat infracost_breakdown_cmd)"
breakdown_output=$(cat infracost_breakdown_cmd | sh)
echo "$breakdown_output" > infracost_breakdown.json

infracost_output_cmd=$(build_output_cmd "infracost_breakdown.json")
echo "$infracost_output_cmd" > infracost_output_cmd
  
echo "Running infracost output using:"
echo "  $ $(cat infracost_output_cmd)"
diff_output=$(cat infracost_output_cmd | sh)

past_total_monthly_cost=$(jq '[.projects[].pastBreakdown.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)
total_monthly_cost=$(jq '[.projects[].breakdown.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)
diff_cost=$(jq '[.projects[].diff.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)

# If both old and new costs are greater than 0
if [ $(echo "$past_total_monthly_cost > 0" | bc -l) = 1 ] && [ $(echo "$total_monthly_cost > 0" | bc -l) = 1 ]; then
  percent=$(echo "scale=6; $total_monthly_cost / $past_total_monthly_cost * 100 - 100" | bc)
fi

# If both old and new costs are less than or equal to 0
if [ $(echo "$past_total_monthly_cost <= 0" | bc -l) = 1 ] && [ $(echo "$total_monthly_cost <= 0" | bc -l) = 1 ]; then
  percent=0
fi

absolute_percent=$(echo $percent | tr -d -)
diff_resources=$(jq '[.projects[].diff.resources[]] | add' infracost_breakdown.json)

if [ "$(echo "$post_condition" | jq '.always')" = "true" ]; then
  echo "Posting comment as post_condition is set to always"
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ "$diff_resources" = "null" ]; then
  echo "Not posting comment as post_condition is set to has_diff but there is no diff"
  cleanup
  exit 0
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ -n "$diff_resources" ]; then
  echo "Posting comment as post_condition is set to has_diff and there is a diff"
elif [ -z "$percent" ]; then
  echo "Posting comment as percentage diff is empty"
elif [ $(echo "$absolute_percent > $percentage_threshold" | bc -l) = 1 ]; then
  echo "Posting comment as percentage diff ($absolute_percent%) is greater than the percentage threshold ($percentage_threshold%)."
else
  echo "Not posting comment as percentage diff ($absolute_percent%) is less than or equal to percentage threshold ($percentage_threshold%)."
  cleanup
  exit 0
fi

if [ ! -z "$GITHUB_ACTIONS" ]; then
  echo "::set-output name=past_total_monthly_cost::$past_total_monthly_cost"
  echo "::set-output name=total_monthly_cost::$total_monthly_cost"
  load_github_env
  post_to_github
elif [ ! -z "$GITLAB_CI" ]; then
  load_gitlab_env
  post_to_gitlab
elif [ ! -z "$CIRCLECI" ]; then
  post_to_circle_ci
elif [ ! -z "$BITBUCKET_PIPELINES" ]; then
  post_to_bitbucket
elif [ ! -z "$SYSTEM_COLLECTIONURI" ]; then
  post_to_azure_devops
fi

if [ ! -z "$SLACK_WEBHOOK_URL" ]; then
  post_to_slack
fi

cleanup
