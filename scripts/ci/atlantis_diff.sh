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
  # Set defaults
  percentage_threshold=${percentage_threshold:-0}
  INFRACOST_BINARY=${INFRACOST_BINARY:-infracost}

  # Export as it's used by infracost, not this script
  export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
  export INFRACOST_CI_ATLANTIS_DIFF=true
}

build_breakdown_cmd () {
  breakdown_cmd="${INFRACOST_BINARY} breakdown --no-color --terraform-plan-file $PLANFILE --format json"

  if [ ! -z "$usage_file" ]; then
    breakdown_cmd="$breakdown_cmd --usage-file $usage_file"
  fi

  if [ "$atlantis_debug" != "true" ]; then
    breakdown_cmd="$breakdown_cmd 2>/dev/null"
  fi

  echo "$breakdown_cmd"
}

build_output_cmd () {
  breakdown_path=$1
  output_cmd="${INFRACOST_BINARY} output --no-color --format=diff $1"
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
  if [ $(echo "$new_monthly_cost < ${old_monthly_cost}" | bc -l) = 1 ]; then
    change_word="decrease"
    change_sym=""
  fi

  percent_display=""
  if [ ! -z "$percent" ]; then
    percent_display=" (${change_sym}${percent}%)"
  fi

  msg="##### Infracost estimate #####"
  msg="${msg}\n\n"
  msg="${msg}Monthly cost will ${change_word} by $(format_cost $diff_cost)$percent_display\n"
  msg="${msg}\n"
  msg="${msg}Previous monthly cost: $(format_cost $old_monthly_cost)\n"
  msg="${msg}New monthly cost: $(format_cost $new_monthly_cost)\n"
  msg="${msg}\n"
  msg="${msg}Infracost output:\n"
  msg="${msg}\n"
  msg="${msg}$(echo "    ${diff_output//$'\n'/\\n    }")\n"
  printf "$msg"
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

old_monthly_cost=$(jq '[.projects[].pastBreakdown.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)
new_monthly_cost=$(jq '[.projects[].breakdown.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)
diff_cost=$(jq '[.projects[].diff.totalMonthlyCost | select (.!=null) | tonumber] | add' infracost_breakdown.json)

# If both old and new costs are greater than 0
if [ $(echo "$old_monthly_cost > 0" | bc -l) = 1 ] && [ $(echo "$new_monthly_cost > 0" | bc -l) = 1 ]; then
  percent=$(echo "scale=4; $new_monthly_cost / $old_monthly_cost * 100 - 100" | bc)
  percent="$(printf "%.0f" $percent)"
fi

# If both old and new costs are less than or equal to 0
if [ $(echo "$old_monthly_cost <= 0" | bc -l) = 1 ] && [ $(echo "$new_monthly_cost <= 0" | bc -l) = 1 ]; then
  percent=0
fi

absolute_percent=$(echo $percent | tr -d -)

if [ -z "$percent" ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Diff percentage is empty"
  fi
elif [ $(echo "$absolute_percent > $percentage_threshold" | bc -l) = 1 ]; then
  if [ "$atlantis_debug" = "true" ]; then
    echo "Diff ($percent%) is greater than the percentage threshold ($percentage_threshold%)."
  fi
else
  if [ "$atlantis_debug" = "true" ]; then
    echo "Comment not posted as diff ($absolute_percent%) is less than or equal to percentage threshold ($percentage_threshold%)."
  fi
  exit 0
fi

msg="$(build_msg)"
echo "$msg"

# Cleanup
rm -f infracost_breakdown_cmd infracost_output_cmd
