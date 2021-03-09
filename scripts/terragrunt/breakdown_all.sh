#!/usr/bin/env bash

# See https://www.infracost.io/docs/iac_tools/terragrunt for usage docs

# Output terraform plans
terragrunt run-all plan -out=infracost-plan

# Loop through plans and output infracost JSONs
planfiles=($(find . -name "infracost-plan" | tr '\n' ' '))
for planfile in "${planfiles[@]}"; do
  echo "Running terraform show for $planfile";
  dir=$(dirname $planfile)
  cd $dir
  terraform show -json $(basename $planfile) > infracost-plan.json
  cd -
  infracost breakdown --path $dir/infracost-plan.json --format json > $dir/infracost-out.json
  rm $planfile
done

# Run infracost output to merge the results
jsonfiles=($(find . -name "infracost-out.json" | tr '\n' ' '))
infracost output --format html $(echo ${jsonfiles[@]/#/--path }) > infracost-report.html
infracost output --format table $(echo ${jsonfiles[@]/#/--path })
echo "Also saved HTML report in infracost-report.html"

# Tidy up
rm $jsonfiles
