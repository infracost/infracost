#!/usr/bin/env bash

# See https://www.infracost.io/docs/terragrunt for usage docs

# Output terraform plans
terragrunt plan-all -out=infracost-plan

# Loop through plans and output infracost JSONs
planfiles=($(find . -name "infracost-plan" | tr '\n' ' '))
for planfile in "${planfiles[@]}"; do
  echo "Running terraform show for $planfile";
  cd $(dirname $planfile)
  terraform show -json $(basename $planfile) > infracost-plan.json
  infracost breakdown --terraform-json-file infracost-plan.json --format json > infracost-out.json
  cd -
done

# Run infracost output to merge the results
jsonfiles=($(find . -name "*-infracost-out.json" | tr '\n' ' '))
infracost output --format html $(echo ${jsonfiles[@]/#/--path }) > infracost-report.html
infracost output --format table $(echo ${jsonfiles[@]/#/--path })
echo "Also saved HTML report in infracost-report.html"

# Tidy up
rm $planfiles
rm $jsonfiles
