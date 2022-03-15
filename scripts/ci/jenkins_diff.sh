#!/bin/sh -le

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# This script is DEPRECATED and is no longer maintained.
#
# Please migrate to our new Jenkins integration: https://github.com/infracost/infracost-jenkins
#
# This script will be removed September 2022.
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
echo "Warning: this script is deprecated and will be removed in Sep 2022."
echo "Please visit https://github.com/infracost/infracost-jenkins/ for the new recommended way of using Infracost in Jenkins."
echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

fix_env_vars () {
  # Jenkins has problems with envs case sensivity
  iac_path=${iac_path:-$IAC_PATH}
  terraform_plan_flags=${terraform_plan_flags:-$TERRAFORM_PLAN_FLAGS}
  terraform_workspace=${terraform_workspace:-$TERRAFORM_WORKSPACE}
  usage_file=${usage_file:-$USAGE_FILE}
  sync_usage_file=${sync_usage_file:-$SYNC_USAGE_FILE}
  config_file=${config_file:-$CONFIG_FILE}
  fail_condition=${fail_condition:-$FAIL_CONDITION}
  show_skipped=${show_skipped:-$SHOW_SKIPPED}
}

process_args () {
  # Validate fail_condition
  if ! echo "$fail_condition" | jq empty; then
    echo "Error: fail_condition contains invalid JSON"
  fi

  # Set defaults
  if [ -n "$fail_condition" ] && [ "$(echo "$fail_condition" | jq '.percentage_threshold')" != "null" ]; then
    fail_percentage_threshold=$(echo "$fail_condition" | jq -r '.percentage_threshold')
  fi
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}

  # Export as it's used by infracost, not this script
  export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
  export INFRACOST_CI_JENKINS_DIFF=true

  if [ -n "$GIT_SSH_KEY" ]; then
    echo "Setting up private Git SSH key so terraform can access your private modules."
    mkdir -p .ssh
    echo "$GIT_SSH_KEY" > .ssh/git_ssh_key
    chmod 600 .ssh/git_ssh_key
    export GIT_SSH_COMMAND="ssh -i $(pwd)/.ssh/git_ssh_key -o 'StrictHostKeyChecking=no'"
  fi
}

build_breakdown_cmd () {
  breakdown_cmd="$INFRACOST_BINARY breakdown --no-color --format json"

  if [ -n "$iac_path" ]; then
    breakdown_cmd="$breakdown_cmd --path $iac_path"
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

build_msg () {
  local percent_display
  local change_word
  local msg

  percent_display=$(percent_display "$past_total_monthly_cost" "$total_monthly_cost" | sed "s/%/%%/g")
  change_word=$(change_word "$past_total_monthly_cost" "$total_monthly_cost")

  msg="\n##### Infracost estimate #####"
  msg="${msg}\n\n"
  msg="${msg}Monthly cost will $change_word by $(format_cost $diff_total_monthly_cost)$percent_display\n"
  msg="${msg}\n"
  msg="${msg}Previous monthly cost: $(format_cost $past_total_monthly_cost)\n"
  msg="${msg}New monthly cost: $(format_cost $total_monthly_cost)\n"
  msg="${msg}\n"
  msg="${msg}Infracost output:\n"
  msg="${msg}\n"
  msg="${msg}$(echo "$diff_output" | sed 's/^/    /' | sed "s/%/%%/g")\n"
  printf "$msg"
}

build_msg_html () {
  # TODO: change it to infracost output once https://github.com/infracost/infracost/issues/509 is resolved.
  msg=$1
  html="<!DOCTYPE html><html><body><pre>"
  html="$html $msg"
  html="$html</pre></body></html>"
  printf '%s' "$html"
}

format_cost () {
  cost=$1

  if [ -z "$cost" ] || [ "$cost" = "null" ]; then
    echo "-"
  elif [ "$(echo "$cost < 100" | bc -l)" = 1 ]; then
    printf "$currency%0.2f" "$cost"
  else
    printf "$currency%0.0f" "$cost"
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
  if [ "$(echo "$new < $old" | bc -l)" = 1 ]; then
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

post_to_slack () {
  echo "Posting comment to Slack"
  msg="$(build_msg false)"
  jq -Mnc --arg msg "$msg" '{"text": "\($msg)"}' | curl -L -X POST -d @- \
    -H "Content-Type: application/json" \
    "$SLACK_WEBHOOK_URL"
}

cleanup () {
  # Don't delete infracost_diff.html here as Jenkinsfile publishes that
  rm -f infracost_breakdown.json infracost_breakdown_cmd infracost_output_cmd
}

# MAIN

fix_env_vars
process_args "$@"

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

absolute_percent=$(echo $percent | tr -d -)
diff_resources=$(jq '[.projects[].diff.resources[]] | add' infracost_breakdown.json)

is_failure=0
if [ -z $percent ]; then
  echo "Passing as percentage diff is empty."
elif [ -z $fail_percentage_threshold ]; then
  echo "Passing as no fail percentage threshold is specified."  
elif [ $(echo "$absolute_percent > $fail_percentage_threshold" | bc -l) = 1 ]; then
  echo "Failing as percentage diff ($absolute_percent%) is greater than the percentage threshold ($fail_percentage_threshold%)."
  is_failure=1
else
  echo "Passing as percentage diff ($absolute_percent%) is less than or equal to percentage threshold ($fail_percentage_threshold%)."
fi

msg="$(build_msg)"
echo "$msg"

html=$(build_msg_html "$msg")
echo "$html" > infracost_diff.html

if [ -n "$SLACK_WEBHOOK_URL" ]; then
  post_to_slack
fi

cleanup

exit $is_failure
