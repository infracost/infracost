#!/usr/bin/env bash

# See https://www.infracost.io/docs/terragrunt for usage docs

# Output terraform plans
terragrunt plan-all -out=infracost-plan

# Loop through plans and output infracost JSONs
planfiles=$(find . -name "infracost-plan")
while IFS= read -r planfile; do
  echo "Running terraform show for $planfile";
  cd $(dirname $planfile)
  terraform show -json $(basename $planfile) > infracost-plan.json
  infracost breakdown --terraform-json-file infracost-plan.json --format json > infracost-out.json
  cd -
done <<< "$planfiles"

# Run infracost output to merge the results
jsonfiles=$(find . -name "infracost-out.json")
infracost output --format=diff $(echo $jsonfiles | tr '\n' ' ')

# Tidy up
rm $(echo $planfiles | tr '\n' ' ')
rm $(echo $jsonfiles | tr '\n' ' ')
