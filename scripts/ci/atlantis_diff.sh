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

# Handle deprecated var names
terraform_plan_flags=${terraform_plan_flags:-$tfflags}

# Set defaults
percentage_threshold=${percentage_threshold:-0}
# Export as it's used by infracost, not this script
export INFRACOST_LOG_LEVEL=${INFRACOST_LOG_LEVEL:-info}
export INFRACOST_CI_ATLANTIS_DIFF=true

infracost_cmd="infracost --no-color --terraform-dir ."
if [ ! -z "$terraform_plan_flags" ]; then
  infracost_cmd="$infracost_cmd --terraform-plan-flags \"$terraform_plan_flags\""
fi
if [ ! -z "$pricing_api_endpoint" ]; then
  infracost_cmd="$infracost_cmd --pricing-api-endpoint $pricing_api_endpoint"
fi
if [ ! -z "$usage_file" ]; then
  infracost_cmd="$infracost_cmd --usage-file $usage_file"
fi
if [ ! -z "$config_file" ]; then
  infracost_cmd="$infracost_cmd --config-file $config_file"
fi
if [ "$atlantis_debug" = "true" ]; then
  echo "$infracost_cmd" > infracost_cmd
else
  echo "$infracost_cmd 2>/dev/null" > infracost_cmd
fi

# Handle Atlantis merge checkout-strategy
current_branch_commit=$(git rev-parse HEAD)
current_branch_previous_commit_email=$(git log -1 --pretty=format:'%ae')
current_branch_previous_commit_message=$(git log -1 --pretty=format:'%B')
if [ "$current_branch_previous_commit_email" = "atlantis@runatlantis.io" ] && [ "$current_branch_previous_commit_message" = "atlantis-merge" ]; then
  if [ "$atlantis_debug" = "true" ]; then echo "Detected Atlantis merge checkout-strategy so checking out head branch to avoid running against Atlantis' temporary merge commit."; fi
  git remote set-branches head $HEAD_BRANCH_NAME &>/dev/null || if [ "$atlantis_debug" = "true" ]; then echo "Could not set-branches $HEAD_BRANCH_NAME, this might prevent switching to it, continuing..."; fi
  git fetch --depth=1 head $HEAD_BRANCH_NAME &>/dev/null || if [ "$atlantis_debug" = "true" ]; then echo "Could not fetch branch head/$HEAD_BRANCH_NAME, no problems, switching to it..."; fi
  # Use 'checkout head/branch' vs the 'switch' that's used in diff.sh to ensure latest branch changes are used locally
  git checkout head/$HEAD_BRANCH_NAME &>/dev/null || (echo "[Infracost] Error: could not switch to branch $HEAD_BRANCH_NAME" && exit 1)
fi

if [ "$atlantis_debug" = "true" ]; then
  echo "Running infracost on current branch using:"
  echo "  $ $(cat infracost_cmd)"
fi
current_branch_output=$(cat infracost_cmd | sh)
# The sed is needed to cause the header line to be different between current_branch_infracost and
# default_branch_infracost, otherwise git diff removes it as its an identical line
echo "$current_branch_output" | sed 's/MONTHLY COST/MONTHLY COST /' > current_branch_infracost.txt
current_branch_monthly_cost=$(cat current_branch_infracost.txt | awk '/OVERALL TOTAL/ { gsub(",",""); printf("%.2f",$NF) }')
if [ "$atlantis_debug" = "true" ]; then echo "current_branch_monthly_cost is $current_branch_monthly_cost"; fi

if [ "$HEAD_BRANCH_NAME" = "$BASE_BRANCH_NAME" ]; then
  if [ "$atlantis_debug" = "true" ]; then echo "Exiting as the current branch was the default branch so nothing more to do."; fi
  exit 0
fi

if [ "$atlantis_debug" = "true" ]; then echo "Switching to default branch"; fi
git remote set-branches origin $BASE_BRANCH_NAME &>/dev/null || if [ "$atlantis_debug" = "true" ]; then echo "Could not set-branches $BASE_BRANCH_NAME, this might prevent switching to it, continuing..."; fi
git fetch --depth=1 origin $BASE_BRANCH_NAME &>/dev/null || if [ "$atlantis_debug" = "true" ]; then echo "Could not fetch branch $BASE_BRANCH_NAME from origin, no problems, switching to it..."; fi
# Use 'checkout origin/branch' vs the 'switch' that's used in diff.sh to ensure latest master changes are used locally
git checkout origin/$BASE_BRANCH_NAME &>/dev/null || (echo "[Infracost] Error: could not switch to branch $BASE_BRANCH_NAME" && exit 1)

if [ "$atlantis_debug" = "true" ]; then git log -n1; fi

# Handle case of new projects in Atlantis where the base branch doesn't have the Terraform files yet
if [ $(find -regex ".*\.\(tf\|hcl\|hcl.json\)" | grep -v .lock.hcl | wc -l) = "0" ]; then
  if [ "$atlantis_debug" = "true" ]; then echo "Default branch does not have this folder, setting its cost to 0."; fi
  default_branch_monthly_cost=0
  touch default_branch_infracost.txt
else
  if [ "$atlantis_debug" = "true" ]; then
    echo "Running infracost on default branch using:"
    echo "  $ $(cat infracost_cmd)"
  fi
  default_branch_output=$(cat infracost_cmd | sh)
  echo "$default_branch_output" > default_branch_infracost.txt
  default_branch_monthly_cost=$(cat default_branch_infracost.txt | awk '/OVERALL TOTAL/ { gsub(",",""); printf("%.2f",$NF) }')
fi

if [ "$atlantis_debug" = "true" ]; then echo "default_branch_monthly_cost is $default_branch_monthly_cost"; fi

# Switch back to try and not confuse Atlantis
git checkout $current_branch_commit &>/dev/null

if [ $(echo "$default_branch_monthly_cost > 0" | bc -l) = 1 ]; then
  percent_diff=$(echo "scale=4; $current_branch_monthly_cost / $default_branch_monthly_cost * 100 - 100" | bc)
else
  if [ "$atlantis_debug" = "true" ]; then echo "Default branch has no cost, setting percent_diff=100 to force a comment"; fi
  percent_diff=100
  # Remove the empty OVERALL TOTAL line to avoid it showing-up in the diff
  sed -i '/OVERALL TOTAL/d' default_branch_infracost.txt
fi
absolute_percent_diff=$(echo $percent_diff | tr -d -)

if [ $(echo "$absolute_percent_diff > $percentage_threshold" | bc -l) = 1 ]; then
  change_word="increase"
  if [ $(echo "$percent_diff < 0" | bc -l) = 1 ]; then
    change_word="decrease"
  fi
  echo "#####"
  echo
  echo "Infracost estimate: monthly cost will $change_word by $absolute_percent_diff% (default branch \$$default_branch_monthly_cost vs current branch \$$current_branch_monthly_cost)"
  echo
  git diff --no-color --no-index default_branch_infracost.txt current_branch_infracost.txt | sed 1,2d | sed 3,5d
else
  if [ "$atlantis_debug" = "true" ]; then echo "Infracost output omitted as default branch and current branch diff ($absolute_percent_diff) is less than or equal to percentage threshold ($percentage_threshold)."; fi
fi
# Cleanup
rm -f infracost_cmd default_branch_infracost.txt current_branch_infracost.txt
