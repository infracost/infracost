# Contributing to Infracost

The overall process for contributing to Infracost is:
1. Check the [project board](https://github.com/infracost/infracost/projects/2) to see if there is something you'd like to work on; these are the issues we'd like to focus on in the near future. There are also [other issues](https://github.com/infracost/infracost/issues) that you might like to check; the issue labels should help you to find a good first issue, or new resources that others have already requested/liked.
2. Create a new issue if there's no issue for what you want to work on. Please put as much as details as you think is necessary, the use-case context is especially helpful if you'd like to receive good feedback.
3. Create a fork, commit and push to your fork. Send a pull request from your fork to this repo with the proposed change (don't forget to run `make lint` and `make fmt` first). Please include unit and integration tests where applicable. We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
4. We'll review your change and provide feedback.

## Development

Install Go dependencies:
```sh
make deps
```

Get an API key.
```sh
make run ARGS="register"
```
Alternatively checkout and run the [cloud-pricing-api](https://github.com/infracost/cloud-pricing-api) and set the `INFRACOST_PRICING_API_ENDPOINT` environment variable to point to it.

Add the API key to your `.env.local` file:
```
cat <<EOF >> .env.local
INFRACOST_API_KEY=XXX
EOF
```

Run the code:
```sh
make run ARGS="--tfdir examples/terraform --usage-file=examples/terraform/infracost-usage.yml"
```

Run all tests:
```sh
make test
```

Exclude integration tests:
```sh
make test ARGS="-v -short"
```

Build:
```sh
make build
```

## Adding new resources

### Quickstart AWS

_Make sure to get familiar with the pricing model of the resource first by reading the cloud vendors pricing page._ To begin, add a new file in `internal/providers/terraform/aws/` as well as an accompanying test file.

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

Finally create a temporary terraform file to test your resource and run (no need to commit that):

```
make run ARGS="--tfdir my_new_terraform/"
```

### Detailed notes

When adding your first resource, we recommend you look at one of the existing resources to see how it's done, for example, check the [nat_gateway.go](internal/providers/terraform/aws/nat_gateway.go) resource. You can then review the [price_explorer](scripts/price_explorer/README.md) scripts that help you find various pricing service filters, and something called a "priceHash" that you need for writing integration tests.

We distinguish the **price** of a resource from its **cost**. Price is the per-unit price advertised by a cloud vendor. The cost of a resource is calculated by multiplying its price by its usage. For example, an EC2 instance might be priced at $0.02 per hour, and if run for 100 hours (its usage), it'll cost $2.00. When adding resources to Infracost, we can always show their price, but if the resource has a usage-based cost component, we can't show its cost. To solve this problem, new resources in Infracost go through two levels of support:

You can add all price components for the resource, even ones that are usage-based, so the price column in the table output is always populated. The hourly and monthly cost for these components will show `-` as illustrated in the following output for AWS Lambda. Once this is done, please send a pull-request to this repo so someone can review/merge it. Try to re-use relevant costComponents from other resources where applicable, e.g. notice how the `newElasticacheResource` function is used in [aws_elasticache_cluster](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/elasticache_cluster.go) and [aws_elasticache_replication_group](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/elasticache_replication_group.go).

Please use [this pull request description](https://github.com/infracost/infracost/pull/91) as a guide on the level of details to include in your PR, including required integration tests.

  ```
  NAME                                        MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.hello_world
  ├─ Requests                                           -  1M requests  0.2000            -             -
  └─ Duration                                           -  GB-seconds    2e-05            -             -
  Total                                                                                   -             -
  ```

#### Adding usage example

Infracost supports passing usage data in through a usage YAML file. When adding a new resource we should add an example of how to specify the usage data in [infracost-usage-example.yml](/infracost-usage-example.yml). This should include an example resource and usage data along with comments detailing what the usage values are. Here's an example of the entry for AWS Lambda:

  ```yaml
  aws_lambda_function.my_function:
    monthly_requests: 100000      # Monthly requests to the Lambda function.
    request_duration_ms: 500      # Average duration of each request in milliseconds.
  ```

When running infracost with `--usage-file path/to/infracost-usage.yml`, Infracost output shows the hourly/monthly cost columns populated with non-zero values:

  ```
  NAME                                        MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.hello_world
  ├─ Requests                                         100  1M requests  0.2000       0.0274       20.0000
  └─ Duration                                   3,750,000  GB-seconds    2e-05       0.0856       62.5001
  Total                                                                              0.1130       82.5001
  ```

### Cost component names and units

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud vendor pricing webpages as users have probably seen those before. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names. Please check [this spreadsheet](https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0) for examples of cost component names and units.

Where a cloud vendor's pricing pages information can be improved for clarify, we'll do that, e.g. on some pricing webpages, AWS mention use "Storage Rate" to describe pricing for "Provisioned IOPS storage", so we use the latter.

#### Resource notes

The following notes are general guidelines, please leave a comment in your pull request if they don't make sense or they can be improved for the resource you're adding.

- count: do not include the count in the cost component name or in brackets. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units:
  - use plural, e.g. hours, months, requests, GB-months, GB (already plural). For a "unit per something", use singular per time unit, e.g. use Per GB per hour. Where it makes sense, instead of "API calls" use "API requests" or "requests" for better consistency.

  - for things where the Terraform resource represents 1 unit, e.g. an `aws_instance`, an `aws_secretsmanager_secret` and a `google_dns_managed_zone`, the units should be months (or hours if that makes more sense). For everything else, the units should be whatever is being charged for, e.g. queries, requests.

  - for data transferred, where you pay for the data per GB, then use `GB`. For storage, where you pay per GB per month, then use `GB-months`. You'll probably see that the Cloud Pricing API's units to use a similar logic. The AWS pricing pages sometimes use a different one than their own pricing API, in that case the pricing API is a better guide.

- unit multiplier: when adding a `costComponent`, set the `UnitMultiplier` to 1 unless the price is for a large number, e.g. set it to `1000000` if the price should be shown "per 1M requests" in the output.

- tiers in names: use the K postfix for thousand, M for million, B for billion and T for trillion, e.g. "Requests (first 300M)" and "Messages (first 1B)". Units should not be included in brackets unless the cost component relates to storage or data transfer, e.g. "Storage (first 1TB)    GB" is more understandable than "Storage (first 1K)    GB" since users understand terabytes and petabytes.

- purchase options: if applicable, include "on-demand" in brackets after the cost component name, e.g. `Database instance (on-demand`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database instance (on-demand, db.t3.medium)`

- storage type: if applicable, include the storage type in brackets in lower case, e.g. `General purpose storage (gp2)`.

- upper/lower case: cost component names should start with a capital letter and use capital letters for acronyms, unless the acronym refers to a type used by the cloud vendor, for example, `General purpose storage (gp2)` (as `gp2` is a type used by AWS) and `Provisioned IOPS storage`.

- unnecessary words: drop the following words from cost component names if the cloud vendor's pricing webpage shows them: "Rate" "Volumes", "SSD", "HDD"

- brackets: only use 1 set of brackets after a component name, e.g. `Database instance (on-demand, db.t3.medium)` and not `Database instance (on-demand) (db.t3.medium)`

- free resources: if there are certain conditions that can be checked inside a resource Go file, which mean there are no cost components for the resource, return a `NoPrice: true` and `IsSkipped: true` response as shown below.
	```
	// Gateway endpoints don't have a cost associated with them
	if vpcEndpointType == "Gateway" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}
	```

- unsupported resources: if there are certain conditions that can be checked inside a resource Go file, which mean that the resource is not yet supported, log a warning to explain what is not supported and return a `nil` response as shown below.
	```
	if d.Get("placement_tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	}
	```

- to conditionally set values based on the Terraform resource values, first check `d.Get("value_name").Type != gjson.Null` like in the [google_container_cluster](https://github.com/infracost/infracost/blob/f7b0594c1ee7d13c0c37acb8edfb36dde223471a/internal/providers/terraform/google/container_cluster.go#L38-L42) resource. In the past we used `.Exists()` but this only checks that the key does not exist in the Terraform JSON, not that the key exists and is set to null.

- use `IgnoreIfMissingPrice: true` if you need to lookup a price in the Cloud Pricing API and NOT add it if there is no price. We use it for EBS Optimized instances since we don't know if they should have that cost component without looking it up.

#### Usage file notes

1. Where possible use similar terminology as the cloud vendor's pricing pages, their cost calculators might also help.

2. Do not prefix things with `average_` as in the future we might want to use nested values, e.g. `request_duration_ms.max`.

3. Use the following units and keep them lower-case:
  - time: ms, secs, mins, hrs, days, weeks, months
  - size: b, kb, mb, gb, tb

4. Put the units last, e.g. `message_size_kb`, `request_duration_ms`.

5. For resources that are continuous in time, do not use prefixes, e.g. use `instances`, `subscriptions`, `storage_gb`. For non-continuous resources, prefix with `monthly_` so users knows what time interval to estimate for, e.g. `monthly_log_lines`, `monthly_requests`.

## Release steps

@alikhajeh1 and @aliscott rotate release responsibilities between them.

1. In [here](https://github.com/infracost/infracost/actions), click on the "Go" build for the master branch, click on Build, expand Test, then use the "Search logs" box to find any line that has "Multiple products found", "No products found for" or "No prices found for". Update the resource files in question to fix these error, often it's because the price filters need to be adjusted to only return 1 result.
2. In the infracost repo, run `git tag vx.y.z && git push origin vx.y.z`
3. Wait for the GH Actions to complete, the [newly created draft release](https://github.com/infracost/infracost/releases/) should have 4 assets.
4. Click on the Edit draft button, add the release notes from the commits between this and the last release and click on publish.
5. Announce the release in the infracost-community Slack general channel. Then wait for the [infracost brew PR](https://github.com/Homebrew/homebrew-core/pulls) to be merged.
6. Update the docs repo with any required changes and supported resources.
7. Close addressed issues and tag anyone who liked/commented in them to tell them it's live in version X.

If a new flag/feature is added that requires CI support, update the 9 repos mentioned [here](https://github.com/infracost/infracost/tree/master/scripts/ci#infracost-ci-scripts). For the GitHub Action, a new tag is needed and the release should be published on the GitHub Marketplace. For the CircleCI orb, the readme mentions the commit prefix that triggers releases to the CircleCI orb marketplace.
