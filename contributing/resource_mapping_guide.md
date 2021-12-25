
# Resource mapping guide

This page documents how we think about naming cost component names and units when mapping resources. You probably don't need to read this page as the Infracost team adds detailed descriptions of cost component names and units in GitHub issues; if you'd like to help with that, please use the guidelines mentioned in this page.

* **Resource** (or Generic resource) - A resource for a cloud service that maps one-to-one to an IaC provider resource (e.g., Terraform).
* **Cost component** - A line-item for a resource representing a cost calculated from a price and some usage quantity. For example, the [AWS NAT Gateway](/internal/resources/aws/nat_gateway.go) resource has two cost components: one to model the hourly cost, and another to model the cost for data processed.

## Cost component names and units

To determine the cost components we can look at the following resources:

- Cloud vendor pricing pages for that service (e.g. https://aws.amazon.com/redshift/pricing/)
- Cloud vendor pricing calculators
  - AWS: https://calculator.aws
  - Google: https://cloud.google.com/products/calculator
  - Azure: https://azure.microsoft.com/en-gb/pricing/calculator/

Our aim is to make Infracost's output understandable without needing to read separate docs. We try to match the cloud vendor pricing webpages as users have probably seen those before. However, we sometimes use clearer names when the cloud vendor's pricing pages are confusing, e.g. on some pricing webpages, AWS uses "Storage Rate" to describe pricing for "Storage (provisioned IOPS SSD)", so we use the latter. It's unlikely that users will have looked at the pricing service JSON (which comes from cloud vendors' pricing APIs), or looked at the detailed billing CSVs that can show the pricing service names. Please check [this spreadsheet](https://docs.google.com/spreadsheets/d/1H_bn2jLzYr7xyrvNsFn-0rDaGPGpnrVTPsjVHzr-kM4/edit#gid=0) for examples of cost component names and units.

The cost component name should be plural where it makes sense, e.g. "Certificate renewals", "Requests", and "Messages". Furthermore, the name should not change when the IaC resource params change; anything that can change should be put in brackets, so for example:
- `General Purpose SSD storage (gp2)` should be `Storage (gp2)` as the storage type can change.
- `Outbound data transfer to EqDC2` should be `Outbound data transfer (to EqDC2)` as the EqDC2 value changes based on the location.
- `Linux/UNIX (on-demand, m1.small)` should be `Instance usage (Linux/UNIX, on-demand, m1.small)`.

In the future, we plan to add a separate field to cost components to hold the metadata in brackets.

## General guidelines

- count: do not include the count in the cost component name or in brackets. Terraform's `count` replicates a resource in `plan.json` file. If something like `desired_count` or other cost-related count parameter is included in the `plan.json` file, do use count when calculating the HourlyQuantity/MonthlyQuantity so each line-item in the Infracost output shows the total price/cost for that line-item.

- units:
  - use plural, e.g. hours, months, requests, GB (already plural). For a "unit per something", use singular per time unit, e.g. use Per GB per hour. Where it makes sense, instead of "API calls" use "API requests" or "requests" for better consistency.

  - for things where the Terraform resource represents 1 unit, e.g. an `aws_instance`, an `aws_secretsmanager_secret` and a `google_dns_managed_zone`, the units should be months (or hours if that makes more sense). For everything else, the units should be whatever is being charged for, e.g. queries, requests.

  - for data transferred where you pay for the data per GB, then use `GB`.

  - for storage or other resources priced in Unit-months (e.g. `GB-months`), then use the unit by itself (`GB`). The AWS pricing pages sometimes use a different unit than their own pricing API, in that case the pricing page is a better guide.

  - for units priced in Unit-hours (e.g. `IOPS-hours`) but best understood in months, then use the unit by itself (`IOPS`) with an appropriate `UnitMultiplier`.

  - unit multiplier: when adding a `costComponent`, set the `UnitMultiplier` to 1 except:

    - If the price is for a large number. E.g. set `Unit: "1M requests", UnitMultiplier: 1000000` if the price should be shown "per 1M requests" in the output.

    - If the price is for billing in Unit-hours but best understood in Unit-months. E.g. set `Unit: "GB", UnitMultiplier: schema.HourToMonthUnitMultiplier` to show "per GB" in the output.

- tiers in names: use the K postfix for thousand, M for million, B for billion and T for trillion, e.g. "Requests (first 300M)" and "Messages (first 1B)". Use the words "first", "next" and "over" when describing tiers. Units should not be included in brackets unless the cost component relates to storage or data transfer, e.g. "Storage (first 1TB)    GB" is more understandable than "Storage (first 1K)    GB" since users understand terabytes and petabytes. You should be able to use the `CalculateTierBuckets` method for calculating tier buckets.

- purchase options: if applicable, include "on-demand" in brackets after the cost component name, e.g. `Database instance (on-demand)`

- instance type: if applicable, include it in brackets as the 2nd argument, after the cost component name, e.g. `Database instance (on-demand, db.t3.medium)`

- storage type: if applicable, include the storage type in brackets in lower case, e.g. `General purpose storage (gp2)`.

- upper/lower case: cost component names should start with a capital letter and use capital letters for acronyms, unless the acronym refers to a type used by the cloud vendor, for example, `General purpose storage (gp2)` (as `gp2` is a type used by AWS) and `Provisioned IOPS storage`.

- unnecessary words: drop the following words from cost component names if the cloud vendor's pricing webpage shows them: "Rate" "Volumes", "SSD", "HDD"

- brackets: only use 1 set of brackets after a component name, e.g. `Database instance (on-demand, db.t3.medium)` and not `Database instance (on-demand) (db.t3.medium)`

## Usage file guidelines

- Where possible use similar terminology as the cloud vendor's pricing pages, their cost calculators might also help.

- Do not prefix things with `average_` as in the future we might want to use nested values, e.g. `request_duration_ms.max`.

- Use the following units and keep them lower-case:
  - time: ms, secs, mins, hrs, days, weeks, months
  - size: b, kb, mb, gb, tb

- Put the units last, e.g. `message_size_kb`, `request_duration_ms`.

- For resources that are continuous in time, do not use prefixes, e.g. use `instances`, `subscriptions`, `storage_gb`. For non-continuous resources, prefix with `monthly_` so users knows what time interval to estimate for, e.g. `monthly_log_lines`, `monthly_requests`.

- When the field accepts a string (e.g. `dx_connection_type: dedicated`), the values should be used in a case-insensitive way in the resource file, the `ValueRegex` option can be used with `/i` to allow case-insensitive regex matches. For example `{Key: "connectionType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", connectionType))},`.
