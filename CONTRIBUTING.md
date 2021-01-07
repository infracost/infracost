# Contributing to Infracost

The overall process for contributing to Infracost is:
1. Check the [project board](https://github.com/infracost/infracost/projects/2) to see if there is something you'd like to work on; these are the issues we'd like to focus on in the near future. There are also [other issues](https://github.com/infracost/infracost/issues) that you might like to check; the issue labels should help you to find a good first issue, or new resources that others have already requested/liked.
2. Create a new issue if there's no issue for what you want to work on. Please put as much as details as you think is necessary, the use-case context is especially helpful if you'd like to receive good feedback.
3. Send a pull request with the proposed change (don't forget to run `make lint` and `make fmt` first). Please include unit and integration tests where applicable. We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
4. We'll review your change and provide feedback.

## Development

Install Go dependencies:
```sh
make deps
```

Install latest version of terraform-provider-infracost. If you want to use a local development version, see [this](#using-a-local-version-of-terraform-provider-infracost)
```sh
make install_provider
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
make run ARGS="--tfdir examples/terraform"
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

If you want to work on the [terraform-provider-infracost repository](https://github.com/infracost/terraform-provider-infracost), you can install the local version in your `~/.terraform.d/plugins` directory by:
```sh
# fork/clone the repo
cd terraform-provider-infracost
make install
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
        instanceCount := 1

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

#### Level 1 support
You can add all price components for the resource, even ones that are usage-based, so the price column in the table output is always populated. The hourly and monthly cost for these components will show `-` as illustrated in the following output for AWS Lambda. Once this is done, please send a pull-request to this repo so someone can review/merge it. Try to re-use relevant costComponents from other resources where applicable, e.g. notice how the `newElasticacheResource` function is used in [aws_elasticache_cluster](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/elasticache_cluster.go) and [aws_elasticache_replication_group](https://github.com/infracost/infracost/blob/master/internal/providers/terraform/aws/elasticache_replication_group.go).

Please use [this pull request description](https://github.com/infracost/infracost/pull/91) as a guide on the level of details to include in your PR, including required integration tests.

  ```
  NAME                                        MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.hello_world
  ├─ Requests                                           -  1M requests  0.2000            -             -
  └─ Duration                                           -  GB-seconds    2e-05            -             -
  Total                                                                                   -             -
  ```

#### Level 2 support
The [Infracost Terraform provider](https://github.com/infracost/terraform-provider-infracost) can be updated to add support for usage-based cost components for the new resource. This enables users to add a new data block into their Terraform file to provide usage estimates as shown below. We recommend reviewing one of the existing resources in the `terraform-provider-infracost` repo to see how it works. Once that's done, please open a pull request against the `terraform-provider-infracost` repo, and one against the `infracost` repo that uses your Terraform provider change to populate the cost fields. **Note** that we're still gathering feedback about the Infracost Terraform provider and we might change approach. We recommend that you create an issue if you'd like to work on this so we can guide you through other options that we might want to consider.

  ```hcl
  # Use the infracost provider to get cost estimates for Lambda requests and duration
  data "infracost_aws_lambda_function" "hello_world_usage" {
    resources = [aws_lambda_function.hello_world.id]

    monthly_requests {
      value = 100000000
    }

    average_request_duration {
      value = 250
    }
  }
  ```

  Infracost output shows the hourly/monthly cost columns populated with non-zero values:

  ```
  NAME                                        MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.hello_world
  ├─ Requests                                         100  1M requests  0.2000       0.0274       20.0000
  └─ Duration                                   3,750,000  GB-seconds    2e-05       0.0856       62.5001
  Total                                                                              0.1130       82.5001
  ```

### Cost component names and units

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud vendor pricing webpages as users have probably seen those before. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names. Please check [this spreadsheet](https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0) for examples of cost component names and units. This spreadsheet is continually updated to add new components based on pull requests and the discussion that goes on inside them. We expect that the spreadsheet will get fewer additions as most cloud vendor resources can re-use similar cost component names/units.

Where a cloud vendor's pricing pages information can be improved for clarify, we'll do that, e.g. on some pricing webpages, AWS mention use "Storage Rate" to describe pricing for "Provisioned IOPS storage", so we use the latter.

**Notes**

The following notes are general guidelines, please leave a comment in your pull request if they don't make sense or they can be improved for the resource you're adding.

- count: do not include the count in the Infracost name. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units: use plural, e.g. hours, requests, GB-months, GB (already plural). For a "unit per something", use singular per time unit, e.g. use Per GB per hour.

- unit multiplier: when adding a `costComponent`, set the `UnitMultiplier` to 1 unless the price is for a large number, e.g. set it to `1000000` if the price should be shown "per 1M requests" in the output.

- tiers in names: use the K postfix for thousand, M for million, B for billion and T for trillion, e.g. "Requests (first 300M)" and "Messages (first 1B)".

- purchase options: if applicable, include "on-demand" in brackets after the cost component name, e.g. `Database instance (on-demand`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database instance (on-demand, db.t3.medium)`

- storage type: if applicable, include the storage type in brackets in lower case, e.g. `General purpose storage (gp2)`.

- upper/lower case: cost component names should start with a capital letter and use capital letters for acronyms, for example, `General purpose storage (gp2)` and `Provisioned IOPS storage`.

- unnecessary words: drop the following words from cost component names if the cloud vendor's pricing webpage shows them: "Rate" "Volumes", "SSD", "HDD"

- brackets: only use 1 set of brackets after a component name, e.g. `Database instance (on-demand, db.t3.medium)` and not `Database instance (on-demand) (db.t3.medium)`

## Releasing steps

1. In the infracost repo, run `git tag vx.y.z && git push origin vx.y.z`
2. Wait for the GH Actions to complete, the [newly created tag](https://github.com/infracost/infracost/releases/latest) should have 6 assets.
3. Click on the Edit release, add the release notes from the commits between this and the last release and click on publish.
4. Announce the release in the infracost-community Slack general channel. Then wait for the [infracost brew PR](https://github.com/Homebrew/homebrew-core/pulls) to be merged.
5. Update the docs repo with any required changes and supported resources.
6. Close addressed issues and tag anyone who liked/commented in them to tell them it's live in version X.
7. If required, bump up `terraform-provider-infracost/version`, commit and `git tag vx.y.z && git push origin vx.y.z` in terraform-provider-infracost repo.
