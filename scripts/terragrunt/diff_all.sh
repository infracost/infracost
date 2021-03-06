#!/usr/bin/env bash

# See https://www.infracost.io/docs/terragrunt for usage docs

# Output terraform plans
terragrunt plan-all -out=infracost-plan

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
infracost output --format=diff $(echo ${jsonfiles[@]/#/--path })

# Tidy up
rm $jsonfiles
