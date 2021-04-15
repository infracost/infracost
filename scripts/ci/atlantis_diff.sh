#!/bin/sh -le

# This is an Atlantis-specific script that runs infracost on the current branch then
# the master branch. It uses `git diff` to output the cost estimate difference
# whenever a percentage threshold is crossed. The output is displayed at the bottom of
# the comments that Atlantis posts on pull requests.
# Usage docs: https://www.infracost.io/docs/integrations/

if [ "$atlantis_debug" = "true" ] || [ "$atlantis_debug" = "True" ] || [ "$atlantis_debug" = "TRUE" ]; then
  atlantis_debug=true
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
  if [ ! -z "$percentage_threshold" ] && [ ! -z "$post_condition" ]; then
    if [ "$atlantis_debug" = "true" ]; then
      echo "Warning: percentage_threshold is deprecated, using post_condition instead"
    fi
  elif [ ! -z "$percentage_threshold" ]; then
    post_condition="{\"percentage_threshold\": $percentage_threshold}"
    if [ "$atlantis_debug" = "true" ]; then
      echo "Warning: percentage_threshold is deprecated and will be removed in v0.9.0, please use post_condition='{\"percentage_threshold\": \"0\"}'"
    fi
  else
    post_condition=${post_condition:-'{"has_diff": true}'}
  fi
  if [ ! -z "$post_condition" ] && [ "$(echo "$post_condition" | jq '.percentage_threshold')" != "null" ]; then
    percentage_threshold=$(echo "$post_condition" | jq -r '.percentage_threshold')
  fi
  percentage_threshold=${percentage_threshold:-0}
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}

  # Export as it's used by infracost, not this script
  export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
  export INFRACOST_CI_ATLANTIS_DIFF=true
}

build_breakdown_cmd () {
  breakdown_cmd="${INFRACOST_BINARY} breakdown --no-color --format json"

  if [ ! -z "$usage_file" ]; then
    breakdown_cmd="$breakdown_cmd --usage-file $usage_file"
  fi
  if [ ! -z "$config_file" ]; then
    breakdown_cmd="$breakdown_cmd --config-file $config_file"
  else
    breakdown_cmd="$breakdown_cmd --path $PLANFILE"
  fi
  if [ "$atlantis_debug" != "true" ]; then
    breakdown_cmd="$breakdown_cmd 2>/dev/null"
  fi

  echo "$breakdown_cmd"
}

build_output_cmd () {
  breakdown_path=$1
  output_cmd="${INFRACOST_BINARY} output --no-color --format diff --path $1"
  echo "${output_cmd}"
}


format_cost () {
  cost=$1

  if [ -z "$cost" ] | [ "${cost}" == "null" ]; then
    echo "-"
  elif [ $(echo "$cost < 100" | bc -l) = 1 ]; then
    printf "$%0.2f" $cost
  else
    printf "$%0.0f" $cost
  fi
}

build_msg () {
  change_word="increase"
  change_sym="+"
  if [ $(echo "$total_monthly_cost < ${past_total_monthly_cost}" | bc -l) = 1 ]; then
    change_word="decrease"
    change_sym=""
  fi

  percent_display=""
  if [ ! -z "$percent" ]; then
    percent_display="$(printf "%.0f" $percent)"
    percent_display=" (${change_sym}${percent_display}%%)"
  fi

  msg="##### Infracost estimate #####"
  msg="${msg}\n\n"
  msg="${msg}Monthly cost will ${change_word} by $(format_cost $diff_cost)$percent_display\n"
  msg="${msg}\n"
  msg="${msg}Previous monthly cost: $(format_cost $past_total_monthly_cost)\n"
  msg="${msg}New monthly cost: $(format_cost $total_monthly_cost)\n"
  msg="${msg}\n"
  msg="${msg}Infracost output:\n"
  msg="${msg}\n"
  msg="${msg}$(echo "    ${diff_output//$'\n'/\\n    }" | sed "s/%/%%/g")\n"
  printf "$msg"
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

cleanup
