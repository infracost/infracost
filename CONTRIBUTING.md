## This page is Work In Progress!

## Adding new resources

When adding your first resource, we recommend you look at one of the existing resources to see how it's done, for example, check the [internal/providers/terraform/aws/nat_gateway.go](nat_gateway.go) resource. You can then review the [price_explorer](scripts/price_explorer/README.md) scripts that help you find various pricing service filters. You can use [this pull request description](https://github.com/infracost/infracost/pull/91) as a guide on the level of detail to include in your PR.

### Cost component names and units

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud provider pricing webpages as users have probably seen those before. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names, however, they are included in the following table to help contributors: https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0

Where AWS' pricing pages information can be improved for clarify, we'll do that, e.g. on some pricing webpages, AWS mention use "Storage Rate" to describe pricing for "Provisioned IOPS Storage", so we use the later.

**Notes**

- count: do not include the count in the Infracost name. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units: use plural, e.g. hours, requests, GB-months. For a "unit per something", use singular per time unit, e.g. use GB/month, not GB/months.

- purchase options: if applicable, include "on_demand" in brackets after the cost component name, e.g. `Database Instance (on demand)`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database Instance (on demand, db.t3.medium)`
