# Contributing to Infracost

🙌 Thank you for contributing and joining our mission to help engineers use cloud infrastructure economically and efficiently 🚀.

## Table of contents

- [Overview](#overview)
- [Setting up the development environment](#setting-up-the-development-environment)
	- [Install](#install)
	- [Run](#run)
	- [Test](#test)
- [Adding new resources](#adding-new-resources)
	- [Glossary](#glossary)
	- [Quickstart](#quickstart)
	- [Cost component names and units](#cost-component-names-and-units)
	- [Finding prices](#finding-prices)
	- [Adding usage-based cost components](#adding-usage-based-cost-components)
	- [General guidelines](#general-guidelines)
	- [Cloud vendor-specific tips](#cloud-vendor-specific-tips)
		- [Google](#google)
		- [Azure](#azure)
- [Releases](#releases)


## Overview

The overall process for contributing to Infracost is:

1. Check the [project board](https://github.com/infracost/infracost/projects/2) to see if there is something you'd like to work on; these are the issues we'd like to focus on in the near future. The issue labels should help you to find an issue to work on. There are also [other issues](https://github.com/infracost/infracost/issues) that you might like to check.
2. Create a new issue if there's no issue for what you want to work on. Please put as much as details as you think is necessary, the use-case context is especially helpful if you'd like to receive good feedback.
3. Add a comment to the issue you're working on to let the rest of the community know.
4. Create a fork, commit and push to your fork. Send a pull request (PR) from your fork to this repo with the proposed change. Don't forget to run `make lint` and `make fmt` first. Please include unit and integration tests where applicable. We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Commit messages can usually start with "feat(aws): add ...", "feat(google): add ...", or "fix: nil pointer...". This helps us generate a cleaner changelog.
5. If it's your first PR to the Infracost org, a bot will leave a comment asking you to follow a quick step to sign our Contributor License Agreement.
6. We'll review your change and provide feedback.

## Setting up the development environment

This guide assumes you are familiar with Terraform, if not you can take an hour to read/watch [this](https://www.terraform.io/intro/index.html) and [this](https://learn.hashicorp.com/collections/terraform/aws-get-started).

### Install

Install go dependencies
```sh
make deps
```

### Run

Run the code:
```sh
make run ARGS="breakdown --path examples/terraform --usage-file=examples/terraform/infracost-usage.yml"
```

This will use your existing [Infracost API key](https://www.infracost.io/docs/#2-get-api-key).

### Test

#### Unit tests

To run only the unit tests:
```sh
make test
```

#### Integration tests

The entire test suite can take >20 mins to run, so we recommend against running them all locally. These will run on GitHub actions.

You should run tests for a file you added/changed with `-v` and warn log level so you can see and fix any warnings:
```sh
INFRACOST_LOG_LEVEL=warn go test -v internal/providers/terraform/aws/ebs_volume_test.go

time="2021-04-05T15:24:16Z" level=warning msg="Multiple prices found for aws_ebs_volume.gp3 Provisioned throughput, using the first price"
```

To run all the tests for a specific cloud vendor:
```sh
make test_aws
make test_google
make test_azure
```

If you do want to run all the tests, you can use:
```sh
make test_all
```

### Build

```sh
make build
```

## Adding new resources

### Glossary

* **Resource** - A resource for a cloud service that maps one-to-one to a Terraform resource.
* **Cost component** - A line-item for a resource that represents a cost calculated from a price and some usage quantity. For example the AWS NAT Gateway resource has 2 cost components: one to model the hourly cost, and another to model the cost for data processed.

### Quickstart

> **Note:** This example uses AWS, but is also applicable to Google and Azure resources.

When adding your first resource, we recommend you first view [this YouTube video](https://www.youtube.com/watch?v=ab7TKRbMlzE). You can also look at one of the existing resources to see how it's done, for example, check the [nat_gateway.go](internal/providers/terraform/aws/nat_gateway.go) resource.

To begin, add a new file in `internal/providers/terraform/aws/` as well as an accompanying test file.

```go
package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetMyResourceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_my_resource",
		RFunc: NewMyResource,
	}
}

func NewMyResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	region := d.Get("region").String()
  var instanceCount int64 = 1

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Instance (on-demand, %s)", "my_instance_type"),
			Unit:           "hours",
			UnitMultiplier:  1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("My AWS Resource family"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Some Usage type/")},
					{Key: "instanceType", Value: strPtr("Some instance type")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
```

Next append the resource to the registry in `internal/providers/terraform/aws/registry.go`.

```go
package aws

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	...,
  GetMyResourceRegistryItem(),
}

```

Finally create a temporary Terraform project, like [this](examples/terraform), to test your resource and run (no need to commit that):

```sh
make run ARGS="breakdown --path my_new_terraform"
```

Please use [this pull request description](https://github.com/infracost/infracost/pull/91) as a guide on the level of details to include in your PR, including required integration tests.

### Cost component names and units

The first thing when adding a new resource is determining what cost components the resource should have. If there is an issue open for the resource then often this information will be included in the issue description (See https://github.com/infracost/infracost/issues/575 for an example). To determine the cost components we can look at the following resources:

- Cloud vendor pricing pages for that service (e.g. https://aws.amazon.com/redshift/pricing/)
- Cloud vendor pricing calculators
	- AWS: https://calculator.aws
	- Google: https://cloud.google.com/products/calculator
	- Azure: https://azure.microsoft.com/en-gb/pricing/calculator/

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud vendor pricing webpages as users have probably seen those before. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names. Please check [this spreadsheet](https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0) for examples of cost component names and units.

Where a cloud vendor's pricing pages information can be improved for clarify, we'll do that, e.g. on some pricing webpages, AWS mention use "Storage Rate" to describe pricing for "Storage (provisioned IOPS SSD)", so we use the latter.

The cost component name should not change when the IaC resource params change; anything that can change should be put in brackets, so for example:
- `General Purpose SSD storage (gp2)` should be `Storage (gp2)` as the storage type can change.
- `Outbound data transfer to EqDC2` should be `Outbound data transfer (to EqDC2)` as the EqDC2 value changes based on the location.
- `Linux/UNIX (on-demand, m1.small)` should be `Instance usage (Linux/UNIX, on-demand, m1.small)`.

In the future, we plan to add a separate field to cost components to hold the metadata in brackets.

### Finding prices

When adding a new resource to infracost, a `ProductFilter` has to be added that uniquely identifies the product for pricing purposes (the filter is used to find a unique `productHash`). If the product contains multiple prices, then a `PriceFilter` should be used to filter down to a single price. The product filter needs a `service`, `productFamily` and `attribute` key/values to filter the products. Because cloud prices have many parameter values to determine a price on a per-region or product type basis, querying for prices can look odd as many parameter values are duplicated.

When writing integration tests we check that we make sure the correct price is used by checking the `priceHash` instead of the actual value of the price. This means that if that price changes, the integration tests won't break. To determine the correct `priceHash` to use for the integration tests we can query the backend data to find a unique product and then look at the prices nested inside the product and choose the correct one based on the price attributes.

Here are two methods for querying the backend data to find the filters that will uniquely identify the product.

#### Querying MongoDB

Instead of directly querying the GraphQL, you can also run `distinct` or `regex` queries on MongoDB to explore the products.

1. Install MongoDB version 4.
2. Download a dump of the MongoDB data from https://infracost-public-dumps.s3.amazonaws.com/cloudPricing.zip.
3. Import the prices to your local:
	```sh
	unzip cloudPricing.zip
	mongorestore -d cloudPricing --gzip cloudPricing
	```

4. You can now query you local MongoDB:

	Run the mongo shell:

	```sh
	mongo
	```

	Example queries:

	```javascript
	use cloudPricing;

	// Find the vendor names
	db.products.distinct("vendorName");
	// should show: [ "aws", "azure", "gcp" ]

	// Find all the services for a vendor
	db.products.distinct("service", {"vendorName": "gcp"})

	// Find the service for a vendor using a regular expression
	db.products.distinct("service", {"vendorName": "gcp", "service": {$regex: "kms", $options: "i"} });
	// should show: [ "Cloud Key Management Service (KMS)", "Thales CPL ekms-dpod-eu" ]

	// Find the product families for a service
	db.products.distinct("productFamily", {"vendorName": "gcp", "service": "Cloud Key Management Service (KMS)"})

	// Find the unique descriptions for a product family
	db.products.distinct("attributes.description", {"vendorName": "gcp", "service": "Cloud Key Management Service (KMS)", "productFamily": "ApplicationServices" });
	// should show:  [
	//	"Active HSM ECDSA P-256 key versions",
	//	"Active HSM ECDSA P-384 key versions",
	//	"Active HSM RSA 2048 bit key versions",
	//  ...
	// ]

	// Find a unique product for a product family based on region and description:
	db.products.find({"vendorName": "gcp", "service": "Cloud Key Management Service (KMS)", "productFamily": "ApplicationServices", "region": "us-east1", "attributes.description": "Active HSM ECDSA P-256 key versions" }).pretty()
	```

#### Querying the GraphQL API

1. Use an browser extension like [modheader](https://bewisse.com/modheader/help/) to allow you to specify additional headers in your browser.
2. Go to https://pricing.api.infracost.io/graphql
3. Set your `X-API-Key` using the browser extension
4. Run GraphQL queries to find the correct products. Examples can be found here: https://github.com/infracost/cloud-pricing-api/tree/master/examples/queries

> **Note:** The GraphQL pricing API limits the number of results returned to 1000, which can limit it's usefulness for exploring the data.

#### Tips

- AWS use many acronyms so be sure to search for those too, e.g. "ES" returns "AmazonES" for ElasticSearch.

### Adding usage-based cost components

We distinguish the **price** of a resource from its **cost**. Price is the per-unit price advertised by a cloud vendor. The cost of a resource is calculated by multiplying its price by its usage. For example, an EC2 instance might be priced at $0.02 per hour, and if run for 100 hours (its usage), it'll cost $2.00. When adding resources to Infracost, we can always show their price, but if the resource has a usage-based cost component, we can't show its cost unless the user specifies the usage.

To do this Infracost supports passing usage data in through a usage YAML file. When adding a new resource we should add an example of how to specify the usage data in [infracost-usage-example.yml](/infracost-usage-example.yml). This should include an example resource and usage data along with comments detailing what the usage values are. Here's an example of the entry for AWS Lambda:

  ```yaml
  aws_lambda_function.my_function:
    monthly_requests: 100000      # Monthly requests to the Lambda function.
    request_duration_ms: 500      # Average duration of each request in milliseconds.
  ```

For an example of a resource with usage-based see the [AWS Lambda resource](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/lambda_function.go). This resource retrieves the quantities from the usage-file by calling `u.Get("monthly_requests")` and `u.Get("request_duration_ms")`. If these quantities are not provided then the `monthlyQuantity` is set to `nil`.

When Infracost is run without usage data the output for this resource looks like:

```
 Name                                                           Quantity  Unit                  Monthly Cost

 aws_lambda_function.hello_world
 ├─ Requests                                            Cost depends on usage: $0.20 per 1M requests
 └─ Duration                                            Cost depends on usage: $0.0000166667 per GB-seconds
```

When Infracost is run with the `--usage-file=path/to/infracost-usage.yml` flag then the output looks like:

```
 Name                                                           Quantity  Unit                  Monthly Cost
 aws_lambda_function.hello_world
 ├─ Requests                                                   100  1M requests        $20.00
 └─ Duration                                            25,000,000  GB-seconds        $416.67
```

### General guidelines

The following notes are general guidelines, please leave a comment in your pull request if they don't make sense or they can be improved for the resource you're adding.

- references to other resources: if you need access to other resources referenced by the resource you're adding, you can specify `ReferenceAttributes`. For example the [aws_ebs_snapshot](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/ebs_snapshot.go#L13) uses this because its price depends on the size of the referenced volume.

- count: do not include the count in the cost component name or in brackets. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units:
  - use plural, e.g. hours, months, requests, GB-months, GB (already plural). For a "unit per something", use singular per time unit, e.g. use Per GB per hour. Where it makes sense, instead of "API calls" use "API requests" or "requests" for better consistency.

  - for things where the Terraform resource represents 1 unit, e.g. an `aws_instance`, an `aws_secretsmanager_secret` and a `google_dns_managed_zone`, the units should be months (or hours if that makes more sense). For everything else, the units should be whatever is being charged for, e.g. queries, requests.

  - for data transferred, where you pay for the data per GB, then use `GB`. For storage, where you pay per GB per month, then use `GB-months`. You'll probably see that the Cloud Pricing API's units to use a similar logic. The AWS pricing pages sometimes use a different one than their own pricing API, in that case the pricing API is a better guide.

- unit multiplier: when adding a `costComponent`, set the `UnitMultiplier` to 1 unless the price is for a large number, e.g. set it to `1000000` if the price should be shown "per 1M requests" in the output.

- tiers in names: use the K postfix for thousand, M for million, B for billion and T for trillion, e.g. "Requests (first 300M)" and "Messages (first 1B)". Use the words "first", "next" and "over" when describing tiers. Units should not be included in brackets unless the cost component relates to storage or data transfer, e.g. "Storage (first 1TB)    GB" is more understandable than "Storage (first 1K)    GB" since users understand terabytes and petabytes. You should be able to use the `CalculateTierBuckets` method for calculating tier buckets.

- purchase options: if applicable, include "on-demand" in brackets after the cost component name, e.g. `Database instance (on-demand`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database instance (on-demand, db.t3.medium)`

- storage type: if applicable, include the storage type in brackets in lower case, e.g. `General purpose storage (gp2)`.

- upper/lower case: cost component names should start with a capital letter and use capital letters for acronyms, unless the acronym refers to a type used by the cloud vendor, for example, `General purpose storage (gp2)` (as `gp2` is a type used by AWS) and `Provisioned IOPS storage`.

- unnecessary words: drop the following words from cost component names if the cloud vendor's pricing webpage shows them: "Rate" "Volumes", "SSD", "HDD"

- brackets: only use 1 set of brackets after a component name, e.g. `Database instance (on-demand, db.t3.medium)` and not `Database instance (on-demand) (db.t3.medium)`

- free resources: if there are certain conditions that can be checked inside a resource Go file, which mean there are no cost components for the resource, return a `NoPrice: true` and `IsSkipped: true` response as shown below.
	```go
	// Gateway endpoints don't have a cost associated with them
	if vpcEndpointType == "Gateway" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}
	```

- unsupported resources: if there are certain conditions that can be checked inside a resource Go file, which mean that the resource is not yet supported, log a warning to explain what is not supported and return a `nil` response as shown below.
	```go
	if d.Get("placement_tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	}
	```

- to conditionally set values based on the Terraform resource values, first check `d.Get("value_name").Type != gjson.Null` like in the [google_container_cluster](https://github.com/infracost/infracost/blob/f7b0594c1ee7d13c0c37acb8edfb36dde223471a/internal/providers/terraform/google/container_cluster.go#L38-L42) resource. In the past we used `.Exists()` but this only checks that the key does not exist in the Terraform JSON, not that the key exists and is set to null.

- use `IgnoreIfMissingPrice: true` if you need to lookup a price in the Cloud Pricing API and NOT add it if there is no price. We use it for EBS Optimized instances since we don't know if they should have that cost component without looking it up.

#### Usage file guidelines

- Where possible use similar terminology as the cloud vendor's pricing pages, their cost calculators might also help.

- Do not prefix things with `average_` as in the future we might want to use nested values, e.g. `request_duration_ms.max`.

- Use the following units and keep them lower-case:
  - time: ms, secs, mins, hrs, days, weeks, months
  - size: b, kb, mb, gb, tb

- Put the units last, e.g. `message_size_kb`, `request_duration_ms`.

- For resources that are continuous in time, do not use prefixes, e.g. use `instances`, `subscriptions`, `storage_gb`. For non-continuous resources, prefix with `monthly_` so users knows what time interval to estimate for, e.g. `monthly_log_lines`, `monthly_requests`.

- When the field accepts a string (e.g. `dx_connection_type: dedicated`), the values should be used in a case-insensitive way in the resource file, the `ValueRegex` option can be used with `/i` to allow case-insensitive regex matches. For example `{Key: "connectionType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", connectionType))},`.

## Cloud vendor-specific tips

### Google

- If the resource has a `zone` key, if they have a zone key, use this logic to get the region:
	```go
	region := d.Get("region").String()
	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}
	```

### Azure

> **Note:** Developing Azure resources requires Azure creds. See below for details.

- The Azure Terraform provider requires real credentials to be able to run `terraform plan`. This means you must have Azure credentials for running the Infracost commands and integration tests for Azure. We recommend creating read-only Azure credentials for this purpose. If you have an Azure subscription, you can do this by running the `az` command line:
	```sh
	az ad sp create-for-rbac --name http://InfracostReadOnly --role Reader --scope=/subscriptions/<SUBSCRIPTION ID> --years=10
	```
	If you do not have an Azure subscription, then please ask on the contributors channel on the Infracost Slack and we can provide you with credentials.

	To run the Azure integration tests in the GitHub action in pull requests, these credentials also need to be added to your fork's secrets. To do this:

	1. Go to `https://github.com/<YOUR GITHUB NAME>/infracost/settings/secrets/actions`.
	2. Add repository secrets for `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID`, `ARM_CLIENT_ID` and `ARM_CLIENT_SECRET`.

## Releases

[@alikhajeh1](https://github.com/alikhajeh1) and [@aliscott](https://github.com/aliscott) rotate release responsibilities between them.

1. In [here](https://github.com/infracost/infracost/actions), click on the "Go" build for the master branch, click on Build, expand Test, then use the "Search logs" box to find any line that has "Multiple products found", "No products found for" or "No prices found for". Update the resource files in question to fix these error, often it's because the price filters need to be adjusted to only return 1 result.
2. In the infracost repo, run `git tag vx.y.z && git push origin vx.y.z`
3. Wait for the GH Actions to complete, the [newly created draft release](https://github.com/infracost/infracost/releases/) should have the darwin-amd64.tar.tz, darwin-arm64.tar.gz, windows-amd64.tar.gz, and linux-amd64.tar.gz assets.
4. Click on the Edit draft button, set the `vx.y.z` value in the tag name and release title. Also add the release notes from the commits between this and the last release and click on publish.
5. In the `infracost-atlantis` repo, run the following steps so the Atlantis integration uses the latest version of Infracost:

	```sh
	# you can also push to master if you want the GH Action to do the following.
	git pull
	docker build --no-cache -t infracost/infracost-atlantis:latest .
	docker push infracost/infracost-atlantis:latest
	```

6. Wait for the [infracost brew PR](https://github.com/Homebrew/homebrew-core/pulls) to be merged.
7. Announce the release in the infracost-community Slack announcements channel.
8. Update the docs repo with any required changes and supported resources.
9. Close addressed issues and tag anyone who liked/commented in them to tell them it's live in version X.

If a new flag/feature is added that requires CI support, update the repos mentioned [here](https://github.com/infracost/infracost/tree/master/scripts/ci#infracost-ci-scripts). For the GitHub Action, a new tag is needed and the release should be published on the GitHub Marketplace. For the CircleCI orb, the readme mentions the commit prefix that triggers releases to the CircleCI orb marketplace.
