#!/bin/bash

terraform init
terraform plan -out=out.plan
terraform show -json out.plan | jq . > plan.json