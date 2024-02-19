# Contributing to Infracost

🙌 Thank you for contributing and joining our mission to help engineers use cloud infrastructure economically and efficiently 🚀.

[Join our community Slack channel](https://www.infracost.io/community-chat), we are a friendly bunch and happy to help you get started :)

## Table of contents

- [Overview](#overview)
- [Setting up the development environment](#setting-up-the-development-environment)
  - [Install](#install)
  - [Run](#run)
  - [Test](#test)
    - [Unit tests](#unit-tests)
    - [Integration tests](#integration-tests)
  - [Build](#build)
- [Adding new resources](#adding-new-resources)
  - [Azure credentials](#azure-credentials)
  - [Querying the GraphQL API](#querying-the-graphql-api)
- [Adding new regions](#adding-new-regions)
  - [AWS](#AWS)
- [Code reviews](#code-reviews)

## Overview

The overall process for contributing to Infracost is:

1. Check the [project board](https://github.com/infracost/infracost/projects/2) to see if there is something you'd like to work on; these are the issues we'd like to focus on in the near future. The issue labels should help you to find an issue to work on. There are also [other issues](https://github.com/infracost/infracost/issues) and [discussions](https://github.com/infracost/infracost/discussions) that you might like to check.
2. Create a new issue if there's no issue for what you want to work on. Please put as much as details as you think is necessary, the use-case context is especially helpful if you'd like to receive good feedback.
3. Add a comment to the issue you're working on to let the rest of the community know.
4. Create a fork, commit and push to your fork. Send a pull request (PR) from your fork to this repo with the proposed change. Don't forget to run `make lint` and `make fmt` first. Please include unit and integration tests where applicable. We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Commit messages can usually start with "feat(aws): add ...", "feat(google): add ...", "fix: nil pointer...", "docs: explain...", or "chore: fix typo". This helps us generate a cleaner changelog.
5. If it's your first PR to the Infracost org, a bot will leave a comment asking you to follow a quick step to sign our Contributor License Agreement.
6. We'll review your change and provide feedback.

## Setting up the development environment

This guide assumes you are familiar with Terraform, if not you can take an hour to read/watch [this](https://www.terraform.io/intro/index.html) and [this](https://learn.hashicorp.com/collections/terraform/aws-get-started).

### Install

Assuming you have already [installed go](https://golang.org/doc/install), install the go dependencies
```sh
make deps
```

### Run

Run the code:
```sh
make run ARGS="breakdown --path examples/terraform --usage-file=examples/terraform/infracost-usage.yml"
```

This will use your existing Infracost API key; register for a [free API key](https://www.infracost.io/docs/#2-get-api-key) key if you don't have one already.

### Test

#### Unit tests

To run only the unit tests:
```sh
make test
```

#### Integration tests
You should run tests with the `-v` flag and warn log level so you can see and fix any warnings:
```sh
INFRACOST_LOG_LEVEL=warn go test -v -cover ./internal/providers/terraform/aws/ebs_volume_test.go

time="2021-04-05T15:24:16Z" level=warning msg="Multiple prices found for aws_ebs_volume.gp3 Provisioned throughput, using the first price"
```

To run all the tests for a specific cloud vendor:
```sh
make test_aws
make test_google
make test_azure
```

To run all the tests, you can use:
```sh
make test_all
```

Test golden files may be updated for all test or for a specific cloud vendor:
```sh
make test_update
make test_update_aws
make test_update_google
make test_update_azure # see the Azure credentials section below
```

### Build

```sh
make build
```

## Adding new resources

Checkout **[our dedicated guide](contributing/add_new_resource_guide.md)** to add resources!

### Azure credentials

Working on Azure resources requires Azure creds as the Azure Terraform provider requires real credentials to be able to run `terraform plan`. This means you must have Azure credentials for running the Infracost commands and integration tests for Azure. We recommend creating read-only Azure credentials for this purpose. If you have an Azure subscription, you can do this by running the `az` command line:
  ```sh
  az ad sp create-for-rbac --name http://InfracostReadOnly --role Reader --scope=/subscriptions/<SUBSCRIPTION ID> --years=10
  ```
  If you do not have an Azure subscription, then please ask on the contributors channel on the Infracost Slack and we can provide you with credentials.

  To run the Azure integration tests in the GitHub action in pull requests, these credentials also need to be added to your fork's secrets. To do this:

  1. Go to `https://github.com/<YOUR GITHUB NAME>/infracost/settings/secrets/actions`.
  2. Add repository secrets for `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID`, `ARM_CLIENT_ID` and `ARM_CLIENT_SECRET`.

### Querying the GraphQL API

1. Use a browser extension like [modheader](https://bewisse.com/modheader/help/) to allow you to specify additional headers in your browser.
2. Go to https://pricing.api.infracost.io/graphql
3. Set your `X-API-Key` using the browser extension
4. Run GraphQL queries to find the correct products. Examples can be found here: https://www.infracost.io/docs/supported_resources/cloud_pricing_api/

The GraphQL pricing API limits the number of results returned to 1000, which can limit its usefulness for exploring the data. AWS use many acronyms so be sure to search for those too, e.g. "ES" returns "AmazonES" for ElasticSearch.

## Adding new regions

### AWS

Consult [PR #2628](https://github.com/infracost/infracost/pull/2628) as an example.

1. In [internal/resources/aws/util.go](internal/resources/aws/util.go), add the new region's information to `RegionMapping`, `RegionCodeMapping`, and `RegionsUsage` as needed.
2. In [internal/resources/aws/global_accelerator_endpoint_group.go](internal/resources/aws/global_accelerator_endpoint_group.go), update `regionCodeMapping` as needed.
3. In [internal/resources/aws/cloudfront_distribution.go](internal/resources/aws/cloudfront_distribution.go), update `regionShieldMapping` as needed.
4. In [internal/providers/terraform/aws/testdata/data_transfer_test/data_transfer_test.usage.yml](internal/providers/terraform/aws/testdata/data_transfer_test/data_transfer_test.usage.yml), add a usage block for data transfer.
5. Update [internal/providers/terraform/aws/testdata/data_transfer_test/data_transfer_test.golden](internal/providers/terraform/aws/testdata/data_transfer_test/data_transfer_test.golden) by running `ARGS="-run TestDataTransferGoldenFile -v -update" make test_aws`.
7. Update [internal/hcl/zones_aws.go]:
   1. Use the AWS CLI to check if the region is enabled in your AWS account by running `aws ec2 describe-availability-zones --region <NEW-REGION_ID e.g. ca-west-1>` 
   2. If needed, enable the region by running `aws account enable-region --region-name <NEW-REGION-ID>`.  This usually takes several minutes. 
   3. From the `internal/hcl` directory, run `go run ../../tools/describezones/main.go aws`.

## Code reviews

Here is a list of things we should look for during code review when adding new resources:

- Is the [infracost-usage-example.yml](https://github.com/infracost/infracost/blob/master/infracost-usage-example.yml) file updated with any new usage file parameters and descriptions?
- Some common bugs that are discovered in reviews:
  - case sensitive string comparisons: `d.Get("kind") ==` should be `strings.ToLower(d.Get("kind").String()) ==`
  - case sensitive regex in price filters: `ValueRegex: strPtr(fmt.Sprintf("/%s/", deviceType))` should be `ValueRegex: strPtr(fmt.Sprintf("/%s/i", deviceType))`
  - missing anchors in price filter regex: `fmt.Sprintf("/%s/", x)` when it should be `fmt.Sprintf("/^%s$/", x)`
  - incorrect output capitalization: └─ Data Ingested should be └─ Data ingested
  - misnamed unit: `GB-month` should be `GB`
- Any "Missing prices" or "Multiple prices" lines when running with `--log-level debug`?
- Any incorrect prices or calculations?
- Any [docs](https://www.infracost.io/docs/) pages need to be updated? e.g. the [supported resources](https://github.com/infracost/docs/blob/master/docs/supported_resources/) pages. If so, please open a PR so it can be merged after the CLI is released.
