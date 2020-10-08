# Contributing to Infracost

The general process for contributing to Infracost is:
1. Check this [Project](https://github.com/infracost/infracost/projects/1) to see if there is something you'd like to work on; these are the ones we'd like to focus on in the near future. There are also [other issues](https://github.com/infracost/infracost/issues) that you might like to check; the issue labels should help you to find a good first issue, or new resources that others have already requested/+1'd.
2. Create a new issue if there's no issue for what you want to work on. Please put as much as details as you think is necessary, the use-case context is especially helpful if you'd like to receive good feedback.
3. Send a pull request with the proposed change (don't forget to run `make lint` and `make fmt` first). Please include unit and integration tests where applicable. We're trialing [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) in case you'd like to use those too.
4. We'll review your change and provide feedback.

## Adding new resources

### Quickstart AWS

_Make sure to get familiar with the pricing model of you resource first._ To begin add a new file in `internal/providers/terraform/aws/` as well as an accompanying test file.

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

func NewMyResource(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	region := d.Get("region").String()
        instanceCount := 1

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Instance (on-demand, %s)", "my_instance_type"),
			Unit:           "hours",
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

Next append the resource to the registry in `internal/providers/terraform/aws/resource_registry.go`.

```go
package aws

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	...,
  GetMyResourceRegistryItem(),
}

```

Finally create a terraform file to test your resource and run:

```
make run ARGS="--tfdir my_new_terraform/"
```

### Overview

When adding your first resource, we recommend you look at one of the existing resources to see how it's done, for example, check the [nat_gateway.go](internal/providers/terraform/aws/nat_gateway.go) resource. You can then review the [price_explorer](scripts/price_explorer/README.md) scripts that help you find various pricing service filters, and something called a "priceHash" that you need for writing integration tests.

We distinguish the **price** of a resource from its **cost**. Price is the per-unit price advertised by a cloud vendor. The cost of a resource is calculated by multiplying its price by its usage. For example, an EC2 instance might be priced at $0.02 per hour, and if run for 100 hours (its usage), it'll cost $2.00. When adding resources to Infracost, we can always show their price, but if the resource has a usage-based cost component, we can't show its cost. To solve this problem, new resources in Infracost go through two levels of support:

1. Level 1 support: you can add all price components for the resource, even ones that are usage-based, so the price column in the table output is always populated. The hourly and monthly cost for these components will show $0.0000 as illustrated in the following output for AWS Lambda. Once this is done, please send a pull-request to this repo so someone can review/merge it. Please use [this pull request description](https://github.com/infracost/infracost/pull/91) as a guide on the level of details to include in your PR, including required integration tests.

  ```
  NAME                              MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.lambda
  ├─ Duration                                 0  GB-seconds    2e-05       0.0000        0.0000
  └─ Requests                                 0  requests      2e-07       0.0000        0.0000
  Total                                                                    0.0000        0.0000
  ```

2. The [Infracost Terraform provider](https://github.com/infracost/terraform-provider-infracost) can be updated to add support for usage-based cost components for the new resource. This enables users to add a new data block into their Terraform file to provide usage estimates as shown below. We recommend reviewing one of the existing resources in the `terraform-provider-infracost` repo to see how it works. Once that's done, please open a pull request against the `terraform-provider-infracost` repo, and one against the `infracost` repo that uses your Terraform provider change to populate the cost fields. **Note** that we're still gathering feedback about the Infracost Terraform provider and we might change approach. We recommend that you create an issue if you'd like to work on this so we can guide you through other options that we might want to consider.

  ```hcl
  # Use the infracost provider to get cost estimates for Lambda requests and duration
  data "infracost_aws_lambda_function" "lambda" {
    resources = [aws_lambda_function.lambda.id]

    monthly_requests {
      value = 100000000
    }

    average_request_duration {
      value = 350
    }
  }
  ```

  Infracost output shows the hourly/monthly cost columns populated with non-zero values:

  ```
  NAME                              MONTHLY QTY  UNIT         PRICE   HOURLY COST  MONTHLY COST

  aws_lambda_function.lambda
  ├─ Duration                          20000000  GB-seconds    2e-05       0.4566      333.3340
  └─ Requests                         100000000  requests      2e-07       0.0274       20.0000
  Total                                                                    0.4840      353.3340
  ```

### Cost component names and units

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud vendor pricing webpages as users have probably seen those before. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names. Please check [this spreadsheet](https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0) for examples of cost component names and units. This spreadsheet is continually updated to add new components based on pull requests and the discussion that goes on inside them. We expect that the spreadsheet will get fewer additions as most cloud vendor resources can re-use similar cost component names/units.

Where a cloud vendor's pricing pages information can be improved for clarify, we'll do that, e.g. on some pricing webpages, AWS mention use "Storage Rate" to describe pricing for "Provisioned IOPS storage", so we use the latter.

**Notes**

The following notes are general guidelines, please leave a comment in your pull request if they don't make sense or they can be improved for the resource you're adding.

- count: do not include the count in the Infracost name. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units: use plural, e.g. hours, requests, GB-months. For a "unit per something", use singular per time unit, e.g. use GB/month, not GB/months.

- purchase options: if applicable, include "on-demand" in brackets after the cost component name, e.g. `Database instance (on-demand`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database instance (on-demand, db.t3.medium)`

- storage type: if applicable, include the storage type in brackets in lower case, e.g. `General purpose storage (gp2)`.

- upper/lower case: cost componet names should start with a capital letter and use capital letters for acronyms, for example, `General purpose storage (gp2)` and `Provisioned IOPS storage`.

- unnecessary words: drop the following words from cost component names if the cloud vendor's pricing webpage shows them: "Rate" "Volumes", "SSD", "HDD"

- brackets: only use 1 set of brackets after a component name, e.g. `Database instance (on-demand, db.t3.medium)` and not `Database instance (on-demand) (db.t3.medium)`
