#!/bin/bash -le

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# This script is DEPRECATED and is no longer maintained.
#
# Please visit : https://www.infracost.io/docs/integrations/cicd/ to migrate to
# to one of our new CI/CD integrations.
#
# This script will be removed September 2022.
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
echo "Warning: this script is deprecated and will be removed in Sep 2022."
echo "Please visit https://github.com/infracost/infracost/blob/master/scripts/ci/README.md for instructions on how to upgrade."
echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

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
  sync_usage_file=${9:-$sync_usage_file}

  # Validate post_condition
  if ! echo "$post_condition" | jq empty; then
    echo "Error: post_condition contains invalid JSON"
  fi

  # Set defaults
  if [ -n "$percentage_threshold" ] && [ -n "$post_condition" ]; then
    echo "Warning: percentage_threshold is deprecated, using post_condition instead"
  elif [ -n "$percentage_threshold" ]; then
    post_condition="{\"percentage_threshold\": $percentage_threshold}"
    echo "Warning: percentage_threshold is deprecated and will be removed in v0.9.0, please use post_condition='{\"percentage_threshold\": \"0\"}'"
  # Default to using update method when posting to GitHub via GitHub actions, Circle CI or Azure DevOps
  # GitHub actions
  elif [ -n "$GITHUB_ACTIONS" ]; then
    post_condition=${post_condition:-'{"update": true}'}
  # CircleCI GitHub
  elif [ -n "$CIRCLECI" ] && echo "$CIRCLE_REPOSITORY_URL" | grep -Eiq github; then
    post_condition=${post_condition:-'{"update": true}'}
  # Azure DevOps GitHub
  elif [ -n "$SYSTEM_COLLECTIONURI" ] && [ "$BUILD_REASON" = "PullRequest" ] && [ "$BUILD_REPOSITORY_PROVIDER" = "GitHub" ]; then
    post_condition=${post_condition:-'{"update": true}'}
  else
    post_condition=${post_condition:-'{"has_diff": true}'}
  fi
  if [ -n "$post_condition" ] && [ "$(echo "$post_condition" | jq '.percentage_threshold')" != "null" ]; then
    percentage_threshold=$(echo "$post_condition" | jq -r '.percentage_threshold')
  fi
  percentage_threshold=${percentage_threshold:-0}
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}
  GITHUB_API_URL=${GITHUB_API_URL:-https://api.github.com}

  # Export as it's used by infracost, not this script
  export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
  export INFRACOST_CI_DIFF=true
  export INFRACOST_CI_POST_CONDITION=$post_condition
  export INFRACOST_CI_PERCENTAGE_THRESHOLD=$percentage_threshold

  if [ -n "$GIT_SSH_KEY" ]; then
    echo "Setting up private Git SSH key so terraform can access your private modules."
    mkdir -p .ssh
    echo "$GIT_SSH_KEY" > .ssh/git_ssh_key
    chmod 600 .ssh/git_ssh_key
    export GIT_SSH_COMMAND="ssh -i $(pwd)/.ssh/git_ssh_key -o 'StrictHostKeyChecking=no'"
  fi

  # Bitbucket Pipelines don't have a unique env so use this to detect it
  if [ -n "$BITBUCKET_BUILD_NUMBER" ]; then
    BITBUCKET_PIPELINES=true
  fi
}

build_breakdown_cmd () {
  breakdown_cmd="$INFRACOST_BINARY breakdown --no-color --format json"

  if [ -n "$path" ]; then
    breakdown_cmd="$breakdown_cmd --path $path"
  fi
  if [ -n "$terraform_plan_flags" ]; then
    breakdown_cmd="$breakdown_cmd --terraform-plan-flags \"$terraform_plan_flags\""
  fi
  if [ -n "$terraform_workspace" ]; then
    breakdown_cmd="$breakdown_cmd --terraform-workspace $terraform_workspace"
  fi
  if [ -n "$usage_file" ]; then
    if [ "$sync_usage_file" = "true" ] || [ "$sync_usage_file" = "True" ] || [ "$sync_usage_file" = "TRUE" ]; then
      breakdown_cmd="$breakdown_cmd --sync-usage-file --usage-file $usage_file"
    else
      breakdown_cmd="$breakdown_cmd --usage-file $usage_file"
    fi
  fi
  if [ -n "$config_file" ]; then
    breakdown_cmd="$breakdown_cmd --config-file $config_file"
  fi
  echo "$breakdown_cmd"
}

build_output_cmd () {
  output_cmd="$INFRACOST_BINARY output --no-color --format diff --path $1"
  if [ -n "$show_skipped" ]; then
    # The "=" is important as otherwise the value of the flag is ignored by the CLI
    output_cmd="$output_cmd --show-skipped=$show_skipped"
  fi
  echo "${output_cmd}"
}

MSG_START="💰 Infracost estimate:"
build_msg () {
  local include_html=$1
  local update_msg=$2

  local percent_display
  local change_word
  local change_emoji
  local msg

  percent_display=$(percent_display "$past_total_monthly_cost" "$total_monthly_cost")
  change_word=$(change_word "$past_total_monthly_cost" "$total_monthly_cost")
  change_emoji=$(change_emoji "$past_total_monthly_cost" "$total_monthly_cost")

  msg="$MSG_START "
  if [ "$diff_total_monthly_cost" != "0" ]; then
    msg+="**monthly cost will $change_word by $(format_cost "${diff_total_monthly_cost#-}")$percent_display** $change_emoji\n"
  else
    msg+="**monthly cost will not change**\n"
  fi
  msg+="\n"

  if [ "$include_html" = true ]; then
    msg+="<table>\n"
    msg+="  <thead>\n"
    msg+="    <td>Project</td>\n"
    msg+="    <td>Previous</td>\n"
    msg+="    <td>New</td>\n"
    msg+="    <td>Diff</td>\n"
    msg+="  </thead>\n"
    msg+="  <tbody>\n"

    local diff_resources
    local skipped_projects
    for (( i = 0; i < project_count; i++ )); do
      diff_resources=$(jq '.projects['"$i"'].diff.resources[]' infracost_breakdown.json)
      if  [ -n "$diff_resources" ] || [ $project_count -eq 1 ]; then
        msg+="$(build_project_row "$i")"
      else
        if [ -z "$skipped_projects" ]; then
          skipped_projects="$(jq -r '.projects['"$i"'].name' infracost_breakdown.json)"
        else
          skipped_projects="$skipped_projects, $(jq -r '.projects['"$i"'].name' infracost_breakdown.json)"
        fi
      fi
    done

    if (( $project_count > 1 )); then
      msg+="$(build_overall_row)"
    fi

    msg+="  </tbody>\n"
    msg+="</table>\n"
    msg+="\n"

    if [ -n "$skipped_projects" ]; then
      msg+="The following projects have no cost estimate changes: $skipped_projects\n\n"
    fi

    msg+="<details>\n"
    msg+="  <summary><strong>Infracost output</strong></summary>\n"
  else
    msg+="Previous monthly cost: $(format_cost "$past_total_monthly_cost")\n"
    msg+="New monthly cost: $(format_cost "$total_monthly_cost")\n"
    msg+="\n"
    msg+="**Infracost output:**\n"
  fi

  msg+="\n"
  msg+="\`\`\`\n"
  msg+="$(echo "$diff_output" | sed "s/%/%%/g")\n"
  msg+="\`\`\`\n"

  if [ "$include_html" = true ]; then
    msg+="</details>\n"
    if [ -n "$update_msg" ]; then
      msg+="\n$update_msg\n\n"
    fi
    msg+="<sub>\n"
    msg+="  Is this comment useful? <a href=\"https://dashboard.infracost.io/feedback/redirect?value=yes\" rel=\"noopener noreferrer\" target=\"_blank\">Yes</a>, <a href=\"https://dashboard.infracost.io/feedback/redirect?value=no\" rel=\"noopener noreferrer\" target=\"_blank\">No</a>, <a href=\"https://dashboard.infracost.io/feedback/redirect?value=other\" rel=\"noopener noreferrer\" target=\"_blank\">Other</a>\n"
    msg+="</sub>\n"
  fi

  printf "$msg"
}

build_project_row () {
  local i=$1

  local max_name_length
  local name
  local label
  local past_monthly_cost
  local monthly_cost
  local diff_monthly_cost
  local percent_display
  local sym

  max_name_length=64
  name=$(jq -r '.projects['"$i"'].name' infracost_breakdown.json)
  # Truncate the middle of the name if it's too long
  name=$(echo $name | awk -v l="$max_name_length" '{if (length($0) > l) {print substr($0, 0, l-(l/2)-1)"..."substr($0, length($0)-(l/2)+3, length($0))} else print $0}')

  past_monthly_cost=$(jq -r '.projects['"$i"'].pastBreakdown.totalMonthlyCost' infracost_breakdown.json)
  monthly_cost=$(jq -r '.projects['"$i"'].breakdown.totalMonthlyCost' infracost_breakdown.json)
  diff_monthly_cost=$(jq -r '.projects['"$i"'].diff.totalMonthlyCost' infracost_breakdown.json)

  if [ "$diff_monthly_cost" != "0" ]; then
    percent_display=$(percent_display "$past_monthly_cost" "$monthly_cost")
  fi

  local row=""
  row+="    <tr>\n"
  row+="      <td>$name</td>\n"
  row+="      <td align=\"right\">$(format_cost "$past_monthly_cost")</td>\n"
  row+="      <td align=\"right\">$(format_cost "$monthly_cost")</td>\n"
  row+="      <td>$(format_cost "$diff_monthly_cost" true)$percent_display</td>\n"
  row+="    </tr>\n"

  printf "%s" "$row"
}

build_overall_row () {
  local percent_display
  local sym

  if [ "$diff_total_monthly_cost" != "0" ]; then
    percent_display=$(percent_display "$past_total_monthly_cost" "$total_monthly_cost")
  fi

  local row=""
  row+="    <tr>\n"
  row+="      <td>All projects</td>\n"
  row+="      <td align=\"right\">$(format_cost "$past_total_monthly_cost")</td>\n"
  row+="      <td align=\"right\">$(format_cost "$total_monthly_cost")</td>\n"
  row+="      <td>$(format_cost "$diff_total_monthly_cost" true)$percent_display</td>\n"
  row+="    </tr>\n"

  printf "%s" "$row"
}

format_cost () {
  cost=$1
  include_plus=$2

  sym=""
  if [ "$(echo "$cost < 0" | bc -l)" = 1 ]; then
    sym="-"
  elif [ "$include_plus" = true ] && [ "$(echo "$cost > 0" | bc -l)" = 1 ]; then
    sym="+"
  fi

  if [ -z "$cost" ] || [ "$cost" = "null" ] || [ "$cost" = "0" ]; then
    cost="0"
  elif [ "$(echo "${cost#-} < 100" | bc -l)" = 1 ]; then
    cost="$(printf "%'0.2f" "$cost")"
  else
    cost="$(printf "%'0.0f" "$cost")"
  fi

  # If the currency length is greater than 1, assume it's a currency code and display `INR -22.78`
  if [ ${#currency} -gt 1 ]; then
    printf "%s" "$currency$sym${cost#-}"
  # If currency length is 1, assume it's a symbol and display it like `-$22.78`
  else
    printf "%s" "$sym$currency${cost#-}"
  fi
}

calculate_percentage () {
  local old=$1
  local new=$2

  local percent=""

  # If both old and new costs are greater than 0
  if [ "$(echo "$old > 0" | bc -l)" = 1 ] && [ "$(echo "$new > 0" | bc -l)" = 1 ]; then
    percent="$(echo "scale=6; $new / $old * 100 - 100" | bc)"
  fi

  # If both old and new costs are less than or equal to 0
  if [ "$(echo "$old <= 0" | bc -l)" = 1 ] && [ "$(echo "$new <= 0" | bc -l)" = 1 ]; then
    percent="0"
  fi

  printf "%s" "$percent"
}

change_emoji () {
  local old=$1
  local new=$2

  local change_emoji="📈"
  if [ "$(echo "$new < $old" | bc -l)" = 1 ]; then
    change_emoji="📉"
  fi

  printf "%s" "$change_emoji"
}

change_word () {
  local old=$1
  local new=$2

  local change_word="increase"
  if [ "$(echo "$new < $old" | bc -l)" = 1 ]; then
    change_word="decrease"
  fi

  printf "%s" "$change_word"
}

change_symbol () {
  local old=$1
  local new=$2

  local change_symbol="+"
  if [ "$(echo "$new <= $old" | bc -l)" = 1 ]; then
    change_symbol=""
  fi

  printf "%s" "$change_symbol"
}

percent_display () {
  local old=$1
  local new=$2

  local percent
  local sym

  percent=$(calculate_percentage "$old" "$new")
  sym=$(change_symbol "$old" "$new")

  local s=""
  if [ -n "$percent" ]; then
    s="$(printf "%.0f" "$percent")"
    s=" ($sym$s%%)"
  fi

  printf "%s" "$s"
}

post_to_github () {
  if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN is required to post comment to GitHub"
  else
    if [ -n "$GITHUB_PULL_REQUEST_NUMBER" ] && [ "$(echo "$post_condition" | jq '.update')" = "true" ]; then
      post_to_github_pull_request
    else
      post_to_github_commit
    fi
  fi
}

post_to_github_commit () {
  echo "Posting comment to GitHub commit $GITHUB_SHA"
  msg="$(build_msg true)"
  jq -Mnc --arg msg "$msg" '{"body": "\($msg)"}' | curl -L -X POST -d @- \
    -H "Content-Type: application/json" \
    -H "Authorization: token $GITHUB_TOKEN" \
    "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/commits/$GITHUB_SHA/comments"
}

fetch_existing_github_pull_request_comments() {
  pull_request_comments="[]" # empty array

  local infra_comments="[]"
  local PER_PAGE=100
  local page=0
  local respLength=0
  while ((page == 0)) || ((respLength == PER_PAGE)); do
    page=$((page+1))

    echo "Fetching comments for pull request $GITHUB_PULL_REQUEST_NUMBER, $page"
    local resp=$(
      curl -L --retry 3 \
      -H "Content-Type: application/json" \
      -H "Authorization: token $GITHUB_TOKEN" \
      "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$GITHUB_PULL_REQUEST_NUMBER/comments?page=$page&per_page=$PER_PAGE"
    )

    infra_comments=${infra_comments},$(echo "${resp}" | jq "[.[] | select(.body | contains(\"${MSG_START}\"))]")

    respLength=$(echo "$resp" | jq length)
  done

  pull_request_comments=$(echo "[$infra_comments]" | jq 'flatten(1)')
}

post_to_github_pull_request () {
  fetch_existing_github_pull_request_comments
  local latest_pr_comment=$(echo "$pull_request_comments" | jq last)

  msg="$(build_msg true "This comment will be updated when the cost estimate changes.")"

  if [ "$latest_pr_comment" != "null" ]; then
    existing_msg=$(echo "$latest_pr_comment" | jq -r .body)
    # '// /' does a string substitution that removes spaces before comparison
    if [ "${msg// /}" != "${existing_msg// /}" ]; then
      local comment_id=$(echo "$latest_pr_comment" | jq -r .id)
      echo "Updating comment $comment_id for pull request $GITHUB_PULL_REQUEST_NUMBER."
      jq -Mnc --arg msg "$msg" '{"body": "\($msg)"}' | curl -L --retry 3 -X PATCH -d @- \
        -H "Content-Type: application/json" \
        -H "Authorization: token $GITHUB_TOKEN" \
        "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/comments/$comment_id"
    else
      echo "Skipping comment for pull request $GITHUB_PULL_REQUEST_NUMBER, no change in msg."
    fi
  else
    echo "Creating new comment for pull request $GITHUB_PULL_REQUEST_NUMBER."
    jq -Mnc --arg msg "$msg" '{"body": "\($msg)"}' | curl -L -X POST -d @- \
      -H "Content-Type: application/json" \
      -H "Authorization: token $GITHUB_TOKEN" \
      "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$GITHUB_PULL_REQUEST_NUMBER/comments"
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
    -u "$BITBUCKET_TOKEN" \
    "https://api.bitbucket.org/2.0/repositories/$1"
}

get_bitbucket_server_url () {
  BITBUCKET_GIT_URL="$(git config --get remote.origin.url)"
  BITBUCKET_PROJECT="$(cut -d '/' -f4 <<<$BITBUCKET_GIT_URL)"
  BITBUCKET_REPO="$(basename $BITBUCKET_GIT_URL .git)"
  printf "https://$BITBUCKET_SERVER_HOSTNAME/rest/api/1.0/projects/$BITBUCKET_PROJECT/repos/$BITBUCKET_REPO/pull-requests"
}

get_bitbucket_server_pr_id () {
  BITBUCKET_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
  curl -sS \
    -H "Authorization: Bearer $BITBUCKET_TOKEN" \
    "$1?at=refs/heads/$BITBUCKET_BRANCH&direction=OUTGOING" \
    | jq -r '.["values"][0].id'
}

post_bitbucket_server_comment () {
  msg="$(build_msg false)"
  echo "Posting comment to $1"
  jq -Mnc --arg msg "$msg" '{"text": "\($msg)"}' | curl -sSL -o /dev/null -X POST -d @- \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $BITBUCKET_TOKEN" \
    "$1"
}

post_to_circle_ci () {
  if echo "$CIRCLE_REPOSITORY_URL" | grep -Eiq github; then
    GITHUB_REPOSITORY="$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME"
    GITHUB_SHA=$CIRCLE_SHA1
    if [ -n "$CIRCLE_PULL_REQUEST" ]; then
      GITHUB_PULL_REQUEST_NUMBER=${CIRCLE_PULL_REQUEST##*/}
      echo "Posting comment from CircleCI to GitHub pull request $GITHUB_PULL_REQUEST_NUMBER"
    else
      echo "Posting comment from CircleCI to GitHub commit $GITHUB_SHA"
    fi
    post_to_github
  elif echo "$CIRCLE_REPOSITORY_URL" | grep -Eiq bitbucket; then
    if [ -n "$CIRCLE_PULL_REQUEST" ]; then
      BITBUCKET_PR_ID=$(echo "$CIRCLE_PULL_REQUEST" | sed 's/.*pull-requests\///')

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
  if [ -n "$BITBUCKET_PR_ID" ]; then
    echo "Posting comment to Bitbucket pull-request $BITBUCKET_PR_ID"
    post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/pullrequests/$BITBUCKET_PR_ID/comments"
  else
    echo "Posting comment to Bitbucket commit $BITBUCKET_COMMIT"
    post_bitbucket_comment "$BITBUCKET_REPO_FULL_NAME/commit/$BITBUCKET_COMMIT/comments"
  fi
}

post_to_bitbucket_server () {
  if [ -z "$BITBUCKET_TOKEN" ]; then
    echo "Error: BITBUCKET_TOKEN is required to post comment to Bitbucket Server"
  else
    BITBUCKET_SERVER_URL="$(get_bitbucket_server_url)"
    BITBUCKET_SERVER_PR_ID="$(get_bitbucket_server_pr_id $BITBUCKET_SERVER_URL)"
    post_bitbucket_server_comment "$BITBUCKET_SERVER_URL/$BITBUCKET_SERVER_PR_ID/comments"
  fi
}

post_to_azure_devops () {
  if [ "$BUILD_REASON" = "PullRequest" ]; then
    if [ "$BUILD_REPOSITORY_PROVIDER" = "GitHub" ]; then
      echo "Posting comment to Azure DevOps GitHub pull-request $SYSTEM_PULLREQUEST_PULLREQUESTNUMBER"
      GITHUB_REPOSITORY=$BUILD_REPOSITORY_NAME
      GITHUB_SHA=$SYSTEM_PULLREQUEST_SOURCECOMMITID
      GITHUB_PULL_REQUEST_NUMBER=$SYSTEM_PULLREQUEST_PULLREQUESTNUMBER
      post_to_github
    elif [ "$BUILD_REPOSITORY_PROVIDER" = "TfsGit" ]; then
      # See https://docs.microsoft.com/en-us/javascript/api/azure-devops-extension-api/commentthreadstatus
      azure_devops_comment_status=${azure_devops_comment_status:-4}
      echo "Posting comment to Azure DevOps repo pull-request $SYSTEM_PULLREQUEST_PULLREQUESTID"
      msg="$(build_msg true)"
      jq -Mnc --arg msg "$msg" --arg azure_devops_comment_status "$azure_devops_comment_status" '{"comments": [{"parentCommentId": 0, "content": "\($msg)", "commentType": 1}], "status": $azure_devops_comment_status}' | curl -L -X POST -d @- \
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

  github_event=$(cat "$GITHUB_EVENT_PATH")

  if [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
    GITHUB_SHA=$(echo "$github_event" | jq -r .pull_request.head.sha)
    GITHUB_PULL_REQUEST_NUMBER=$(echo "$github_event" | jq -r .pull_request.number)
    VCS_PULL_REQUEST_URL=$(echo "$github_event" | jq -r .pull_request.html_url)
    export VCS_PULL_REQUEST_URL
  else
    VCS_PULL_REQUEST_URL=$(curl -s \
      -H "Accept: application/vnd.github.groot-preview+json" \
      -H "Authorization: token $GITHUB_TOKEN" \
      "$GITHUB_API_URL"/repos/"$GITHUB_REPOSITORY"/commits/"$GITHUB_SHA"/pulls \
      | jq -r '. | map(select(.state == "open")) | . |= sort_by(.updated_at) | reverse | .[0].html_url')
    export VCS_PULL_REQUEST_URL
  fi
}

load_gitlab_env () {
  export VCS_REPO_URL=$CI_REPOSITORY_URL

  first_mr=$(echo "$CI_OPEN_MERGE_REQUESTS" | cut -d',' -f1)
  repo=$(echo "$first_mr" | cut -d'!' -f1)
  mr_number=$(echo "$first_mr" | cut -d'!' -f2)
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
if [ -n "$GITHUB_ACTIONS" ]; then
  load_github_env
elif [ -n "$GITLAB_CI" ]; then
  load_gitlab_env
elif [ -n "$CIRCLECI" ]; then
  load_circle_ci_env
elif [ -n "$SYSTEM_COLLECTIONURI" ]; then
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

project_count=$(jq -r '.projects | length' infracost_breakdown.json)
past_total_monthly_cost=$(jq '(.pastTotalMonthlyCost // 0) | tonumber' infracost_breakdown.json)
total_monthly_cost=$(jq '(.totalMonthlyCost // 0) | tonumber' infracost_breakdown.json)
diff_total_monthly_cost=$(jq '(.diffTotalMonthlyCost // 0) | tonumber' infracost_breakdown.json)
currency=$(jq -r '.currency | select (.!=null)' infracost_breakdown.json)
if [ "$currency" = "" ] || [ "$currency" = "USD" ]; then
  currency="$"
elif [ "$currency" = "EUR" ]; then
  currency="€"
elif [ "$currency" = "GBP" ]; then
  currency="£"
else
  currency="$currency " # Space is needed so output is "INR 123"
fi

percent=$(calculate_percentage "$past_total_monthly_cost" "$total_monthly_cost")

absolute_percent=$(echo "$percent" | tr -d -)
diff_resources=$(jq '[.projects[].diff.resources[]] | add' infracost_breakdown.json)

if [ "$(echo "$post_condition" | jq '.always')" = "true" ]; then
  echo "Posting comment as post_condition is set to always"
elif [ "$(echo "$post_condition" | jq '.update')" = "true" ]; then
  if [ "$diff_resources" = "null" ]; then
    echo "Not posting comment as post_condition is set to update but there is no diff"
    cleanup
    exit 0
  else
    echo "Posting comment as post_condition is set to update and there is a diff"
  fi
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ "$diff_resources" = "null" ]; then
  echo "Not posting comment as post_condition is set to has_diff but there is no diff"
  cleanup
  exit 0
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ -n "$diff_resources" ]; then
  echo "Posting comment as post_condition is set to has_diff and there is a diff"
elif [ -z "$percent" ]; then
  echo "Posting comment as percentage diff is empty"
elif [ "$(echo "$absolute_percent > $percentage_threshold" | bc -l)" = 1 ]; then
  echo "Posting comment as percentage diff ($absolute_percent%) is greater than the percentage threshold ($percentage_threshold%)."
else
  echo "Not posting comment as percentage diff ($absolute_percent%) is less than or equal to percentage threshold ($percentage_threshold%)."
  cleanup
  exit 0
fi

if [ -n "$GITHUB_ACTIONS" ]; then
  echo "::set-output name=past_total_monthly_cost::$past_total_monthly_cost"
  echo "::set-output name=total_monthly_cost::$total_monthly_cost"
  post_to_github
elif [ -n "$GITLAB_CI" ]; then
  post_to_gitlab
elif [ -n "$CIRCLECI" ]; then
  post_to_circle_ci
elif [ -n "$BITBUCKET_PIPELINES" ]; then
  post_to_bitbucket
elif [ -n "$BITBUCKET_SERVER_HOSTNAME" ]; then
  post_to_bitbucket_server
elif [ -n "$SYSTEM_COLLECTIONURI" ]; then
  post_to_azure_devops
fi

if [ -n "$SLACK_WEBHOOK_URL" ]; then
  post_to_slack
fi

cleanup
