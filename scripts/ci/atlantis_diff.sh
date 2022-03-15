#!/bin/sh -le

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# This script is DEPRECATED and is no longer maintained.
#
# Please follow our guide: https://www.infracost.io/docs/guides/atlantis_migration/ to migrate
# to our new Atlantis integration: https://github.com/infracost/infracost-atlantis
#
# This script will be removed September 2022.
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

if [ "$atlantis_debug" = "true" ] || [ "$atlantis_debug" = "True" ] || [ "$atlantis_debug" = "TRUE" ]; then
  atlantis_debug=true
  echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
  echo "Warning: this script is deprecated and will be removed in Sep 2022."
  echo "Please visit https://www.infracost.io/docs/guides/atlantis_migration/ for instructions on how to upgrade."
  echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
  echo

  echo "#####"
  echo "Running Infracost atlantis_diff.sh in debug mode, remove atlantis_debug=true from Atlantis configs to disable debug outputs."
  echo
fi

process_args () {
  # Validate post_condition
  if ! echo "$post_condition" | jq empty; then
    echo "Error: post_condition contains invalid JSON"
  fi

  # Set defaults
  if [ -n "$percentage_threshold" ] && [ -n "$post_condition" ]; then
    if [ "$atlantis_debug" = "true" ]; then
      echo "Warning: percentage_threshold is deprecated, using post_condition instead"
    fi
  elif [ -n "$percentage_threshold" ]; then
    post_condition="{\"percentage_threshold\": $percentage_threshold}"
    if [ "$atlantis_debug" = "true" ]; then
      echo "Warning: percentage_threshold is deprecated and will be removed in v0.9.0, please use post_condition='{\"percentage_threshold\": \"0\"}'"
    fi
  else
    post_condition=${post_condition:-'{"has_diff": true}'}
  fi
  if [ -n "$post_condition" ] && [ "$(echo "$post_condition" | jq '.percentage_threshold')" != "null" ]; then
    percentage_threshold=$(echo "$post_condition" | jq -r '.percentage_threshold')
  fi
  percentage_threshold=${percentage_threshold:-0}
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}

  # Export as it's used by infracost, not this script
  export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
  export INFRACOST_CI_ATLANTIS_DIFF=true
}

build_breakdown_cmd () {
  breakdown_cmd="$INFRACOST_BINARY breakdown --no-color --format json"

  if [ -n "$usage_file" ]; then
    if [ "$sync_usage_file" = "true" ] || [ "$sync_usage_file" = "True" ] || [ "$sync_usage_file" = "TRUE" ]; then
      breakdown_cmd="$breakdown_cmd --sync-usage-file --usage-file $usage_file"
    else
      breakdown_cmd="$breakdown_cmd --usage-file $usage_file"
    fi
  fi
  if [ -n "$config_file" ]; then
    breakdown_cmd="$breakdown_cmd --config-file $config_file"
  else
    if [ -f "$PLANFILE.json" ]; then
      breakdown_cmd="$breakdown_cmd --path $PLANFILE.json"
    elif [ -f "$PLANFILE" ]; then
      breakdown_cmd="$breakdown_cmd --path $PLANFILE"
    else
      echo "Error: PLANFILE '$PLANFILE' does not exist"
    fi
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

  msg="##### Infracost estimate #####"
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
  msg="$(build_msg false)"
  if [ "$atlantis_debug" = "true" ]; then
    echo "Posting comment to Slack"
    jq -Mnc --arg msg "$msg" '{"text": "\($msg)"}' | curl -L -X POST -d @- \
      -H "Content-Type: application/json" \
      "$SLACK_WEBHOOK_URL"
  else
    jq -Mnc --arg msg "$msg" '{"text": "\($msg)"}' | curl -S -s -o /dev/null -L -X POST -d @- \
      -H "Content-Type: application/json" \
      "$SLACK_WEBHOOK_URL"
  fi
}

cleanup () {
  rm -f infracost_breakdown.json infracost_breakdown_cmd infracost_output_cmd
}

# MAIN

process_args "$@"

infracost_breakdown_cmd=$(build_breakdown_cmd)
echo "$infracost_breakdown_cmd" > infracost_breakdown_cmd

if [ "$atlantis_debug" = "true" ]; then
  echo "Running infracost breakdown using:"
  echo "  $ $(cat infracost_breakdown_cmd)"
fi
breakdown_output=$(cat infracost_breakdown_cmd | sh)
echo "$breakdown_output" > infracost_breakdown.json

infracost_output_cmd=$(build_output_cmd "infracost_breakdown.json")
echo "$infracost_output_cmd" > infracost_output_cmd

if [ "$atlantis_debug" = "true" ]; then
  echo "Running infracost output using:"
  echo "  $ $(cat infracost_output_cmd)"
fi
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

if [ "$(echo "$post_condition" | jq '.always')" = "true" ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Posting comment as post_condition is set to always"
  fi
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ "$diff_resources" = "null" ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Not posting comment as post_condition is set to has_diff but there is no diff"
  fi
  cleanup
  exit 0
elif [ "$(echo "$post_condition" | jq '.has_diff')" = "true" ] && [ -n "$diff_resources" ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Posting comment as post_condition is set to has_diff and there is a diff"
  fi
elif [ -z "$percent" ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Posting comment as percentage diff is empty"
  fi
elif [ $(echo "$absolute_percent > $percentage_threshold" | bc -l) = 1 ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Posting comment as percentage diff ($absolute_percent%) is greater than the percentage threshold ($percentage_threshold%)."
  fi
else
  if [ "$atlantis_debug" = "true" ]; then
    echo "Not posting comment as percentage diff ($absolute_percent%) is less than or equal to percentage threshold ($percentage_threshold%)."
  fi
  cleanup
  exit 0
fi

msg="$(build_msg)"
echo "$msg"

if [ -n "$SLACK_WEBHOOK_URL" ]; then
  post_to_slack
fi

cleanup
