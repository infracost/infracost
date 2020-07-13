#!/bin/sh -l

output=$(infracost --no-color --tfdir /github/workspace/$1/$TERRAFORM_DIR)
echo "$output"
echo "$output" > $1-infracost.txt

monthly_cost=$(echo $output | awk '/OVERALL TOTAL/ { print $NF }')
echo "::set-output name=monthly_cost::$monthly_cost"
