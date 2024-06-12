# Adding a new resource to Infracost

This guide will help you master the art of adding cloud resources to Infracost. It might seem long but it's pretty easy to follow, we've even built tooling to auto-generate the required scaffolding and there are integration tests too!

Adding new resources is hugely valuable to the community as once a mapping has been coded, everyone can get a quick cost estimate for it. Every new resource mapping gets users closer to loving their cloud bill! 💰📉

## Overview

We're going to explain how Infracost manages cloud resources as that helps with adding support for new resources.

This guide assumes that a GitHub issue exists with a detailed description of the resource's cost components and units that should be added. Read [this guide](resource_mapping_guide.md) if you'd like to learn how to prepare such details and rationale behind the choice of names/units/etc.

The steps below describe adding an AWS resource using Terraform IaC provider, but they are similar for Google and Azure resources. You can follow along, replacing the example values with your resource's information.

## Table of contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Glossary](#glossary)
- [GitHub issue](#github-issue)
- [Generate resource files](#generate-resource-files)
- [Run tests](#run-tests)
- [Add cost components](#add-cost-components)
  - [Cost component structure](#cost-component-structure)
  - [Price search](#price-search)
  - [Resource attributes mapping](#resource-attributes-mapping)
  - [Add usage-based cost components](#add-usage-based-cost-components)
- [Add free resources](#add-free-resources)
- [Update Infracost docs](#update-infracost-docs)
- [Prepare a Pull Request for review](#prepare-a-pull-request-for-review)
- [Appendix](#appendix)
  - [Resource edge cases](#resource-edge-cases)
  - [References to other resources](#references-to-other-resources)
  - [Google zone mappings](#google-zone-mappings)
  - [Azure zone mappings](#azure-zone-mappings)
  - [Region usage](#region-usage)

## Prerequisites

To follow this guide, you will need:

- A working development environment for the Infracost CLI. You can set it up using [this guide](/CONTRIBUTING.md#setting-up-the-development-environment).
- Access to the Cloud Pricing API database via PostgreSQL. Contact us in the [Slack channel](https://www.infracost.io/community) if you need access.

## Glossary

* **Resource** (or Generic resource) - A resource for a cloud service that maps one-to-one to an IaC provider resource (e.g., Terraform).
* **Cost component** - A line-item for a resource representing a cost calculated from a price and some usage quantity. For example, the [AWS NAT Gateway](/internal/resources/aws/nat_gateway.go) resource has two cost components: one to model the hourly cost, and another to model the cost for data processed.

## GitHub issue

We are going to add resources for `AWS Transfer Family` service. The [GitHub issue](https://github.com/infracost/infracost/issues/1030) includes:

- The resource we're going to add
- List of cost components the resource should have
- Links to AWS/Terraform websites for more information.
- A list of free resources that go along with the main resource, if any.

Leave a comment that you're going to work on this issue, and let's begin!

## Generate resource files

Infracost CLI has a dedicated command that makes generating file structure for new resources quick and easy.

According to the GitHub issue, the cloud provider is `aws`, and the resource name should be `transfer_server`. You can notice a naming pattern as Terraform's resource is called `aws_transfer_server`.

Start by running the following command inside the project's directory:

```sh
go run ./cmd/resourcegen/main.go -cloud-provider aws -resource-name transfer_server
```

> **Info**: This command also supports `-with-help true` flag that generates additional code examples and helpful comments. It is set `false` by default.

The command outputs the following:

```sh
Successfully generated resource aws_transfer_server, files written:

        internal/providers/terraform/aws/transfer_server.go
        internal/providers/terraform/aws/transfer_server_test.go
        internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.golden
        internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.tf
        internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.usage.yml
        internal/resources/aws/transfer_server.go

Added function getTransferServerRegistryItem to resource registry:

        internal/providers/terraform/aws/registry.go

Start by adding an example resource to the Terraform test file:

        internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.tf

and running the following command to generate initial Infracost output:

        ARGS="--run TestTransferServer -v -update" make test_aws

Check out 'contributing/add_new_resource_guide.md' guide for next steps!

Happy hacking!!
```

The resource file structure is the following:

```
└─ internal
  ├─ providers
  │ └─ terraform
  │   └─ aws
  │     ├─ <my_resource>.go                   # * Mappings from the Terraform attributes to the generic resource
  │     ├─ <my_resource>_test.go              # Integration tests for the resource
  │     ├─ testdata
  │     │  └─ <my_resource>_test
  │     │    ├─ <my_resource>_test.golden     # Expected output from the golden test
  │     │    ├─ <my_resource>_test.tf         # * Terraform code for running the golden test
  │     │    └─ <my_resource>_test.usage.yml  # * Any usage data for running the golden test
  │     └─ registry.go                        # * List of provider's supported resources
  └─ resources
    └─ aws
      └─ <my_resource>.go                     # * Generic resource that contains the mappings to cost components
```

We are going to edit the files marked with the asterisk.

> Note: Go's naming conventions require keeping acronyms uppercased. The resourcegen command doesn't do that automatically. If your resource includes an abbreviation, you'll have to update all the places manually it's being used. For example, `MwaaEnvironment` should become `MWAAEnvironment`.

## Run tests

Infracost CLI has a test suite. Each resource also has tests. We will run them to ensure that the code we're adding is working correctly.

As the resourcegen command's output suggests, we can run our newly generated tests and see what happens. Most likely, they will fail as cloud resources require specific attributes, and Terraform will throw an error if they are missing. The IaC provider's resource page may help find an example with such details. For `aws_transfer_server`, we can use an example resource here [Basic usage](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/transfer_server#basic).

Here's how Terraform's test file `internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.tf` looks like:

```tf
provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for TransferServer below

# resource "aws_transfer_server" "transfer_server" {
#  # add required attributes
# }
```

We may need to test several cases. A separate resource can represent each test case.

It already includes a required `provider` block. We can paste our example resource under the comment line and save the file. It will look like this:

```tf
resource "aws_transfer_server" "example" {
  tags = {
    Name = "Example"
  }
}
```

Let's run our tests:

```sh
ARGS="--run TestTransferServer -v -update" make test_aws
```

The output should look similar to this:

```
INFRACOST_LOG_LEVEL=warn go test -timeout 30m -ldflags="-X 'github.com/infracost/infracost/internal/version.Version=v0.9.14+13-gec7fae1b-dirty'" ./internal/providers/terraform/aws --run TestTransferServer -v -update
time="2021-12-02T10:51:17+01:00" level=info msg="Running command: /Users/infracost/.asdf/shims/terraform init -no-color"
=== RUN   TestTransferServerGoldenFile
=== PAUSE TestTransferServerGoldenFile
=== CONT  TestTransferServerGoldenFile
    testutil.go:180: Wrote golden file testdata/transfer_server_test/transfer_server_test.golden
--- PASS: TestTransferServer (5.63s)
PASS
ok      github.com/infracost/infracost/internal/providers/terraform/aws       7.130s
```

The `PASS` message tells us that it worked successfully. The provided `-update` flag wrote the test output to the golden file. If we open `internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.golden` file we should see this:

```
 Name  Monthly Qty  Unit  Monthly Cost

 OVERALL TOTAL                   $0.00
──────────────────────────────────
1 cloud resource was detected:
∙ 1 was estimated, 1 includes usage-based costs, see https://infracost.io/usage-file
```

That's a great start! Our generated code works. It doesn't show much yet, but it should show what has changed after adding more code and rerunning the tests. Pretty neat!

With tests in place, we can add our first cost component now!

## Add cost components

The first cost component is usually the hardest, but an important one as it requires most research work: we need to understand _how_ to add it to the resource and _how_ to find its pricing in Pricing API.

> Note: The GitHub issue should already include information about cost components names, units, and maybe even some pricing details. However, sometimes some details may be missing or incomplete. If this happens, we encourage you to ask for clarifications in the issue comments.

We can divide the process of adding a cost component into the following steps:

1. Define a new cost component in the resource's file.
1. Find the product records and prices in the Cloud Pricing API database.
1. If the resource relies on Terraform's attributes, add their mapping to the provider resource file.
1. If a cost component is usage-based, define the resource's usage schema and attributes.

The first three components in the [GitHub issue](https://github.com/infracost/infracost/issues/1030) are `FTP protocol enabled`, `SFTP protocol enabled` and `FTPS protocol enabled`. They all rely on the `protocols` Terraform attribute, so they probably will use the same cost component. We can start with them.

The main resource's file is located in `internal/resources/<cloud provider>' directory` - `internal/resources/aws/transfer_server.go`:

```go
package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// TransferServer struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/<PATH/TO/RESOURCE>/
// Pricing information: https://aws.amazon.com/<PATH/TO/PRICING>/
type TransferServer struct {
	Address string
	Region  string
}

// TransferServerUsageSchema defines a list which represents the usage schema of TransferServer.
var TransferServerUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the TransferServer.
// It uses the `infracost_usage` struct tags to populate data into the TransferServer.
func (r *TransferServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid TransferServer struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		// TODO: add cost components
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    TransferServerUsageSchema,
		CostComponents: costComponents,
	}
}
```

### Cost component structure

The `BuildResource()` function defines the resource's cost components. Let's edit the file with a cost component scaffold code. We'll use the "FTP" protocol as an example and update it later for other protocols:

```go
func (r *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.protocolEnabledCostComponent(),
	}
	...
}

func (r *TransferServer) protocolEnabledCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "FTP protocol enabled",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("<SERVICE>"),
			ProductFamily: strPtr("<PRODUCT FAMILY>"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "<ATTRIBUTE KEY>", Value: strPtr("<STRING FILTER>")},
				{Key: "<ATTRIBUTE KEY>", ValueRegex: regexPtr("<REGEX FILTER>")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
```

Here are the main parts of the cost component:

- `Name` is a short descriptive name of the component.
- `Unit` is the unit of resource component's measurement. For example, it can be `hours` or `10M requests`.
`UnitMultiplier` is used to calculate the cost of component quantity correctly. For example, if a price is `$0.02 per 1k requests`, and assuming the amount is 10,000, its cost will be calculated as `quantity/unitMultiplier * price`.
- `HourlyQuantity` or `MonthlyQuantity` attributes specify the quantity of the resource. If the measurement unit is `GB`, it will be the number of gigabytes. If the unit is `hours`, it can be `1` as "1 hour".
`ProductFilter` is a struct that helps uniquely identify components' "product" in the Pricing API database. It uses the information we find from our [price search](#price-search) step.
- `PriceFilter` helps identify the exact price of the "product." Usually, it's only one, but if there are pricing tiers, its filters can pick the correct value.

We already can populate some placeholders like names and units from the GitHub issue.

We can rerun the tests to see what this gives us. Notice that we don't use `-update` flag anymore to be able to see the diff:

```sh
ARGS="--run TestTransferServer -v" make test_aws
```

```
        Output does not match golden file:

        --- Expected
        +++ Actual
        @@ -1,5 +1,8 @@

        - Name  Monthly Qty  Unit  Monthly Cost
        -
        - OVERALL TOTAL                   $0.00
        + Name                                 Monthly Qty  Unit   Monthly Cost
        +
        + aws_transfer_server.transfer_server
        + └─ FTP protocol enabled                      730  hours         $0.00
        +
        + OVERALL TOTAL                                                   $0.00
         ──────────────────────────────────

--- FAIL: TestTransferServerGoldenFile (1.93s)
FAIL
FAIL    github.com/infracost/infracost/internal/providers/terraform/aws 3.335s
FAIL
make: *** [test_aws] Error 1
```

The test failed, but it gave us helpful information:

- The resource detects the cost component.
- It shows its monthly quantity in hours (730 hours in one month).
- Its cost is $0.

The code works! Our next step is finding the correct price now.

Let's find out how to find the component representation in the Pricing API database.

### Price search

We'll use the PostgreSQL database and `psql` tool to find the prices for cost components.

Connect to `cloud_pricing` database that holds all the prices:

```sh
psql cloud_pricing
```

Infracost stores over 3M prices for multiple services, and each cloud provider may have a different structure for storing such data. JSON is a great format to keep such information. However, it might be a bit overwhelming to search through.

Our goal is to identify the filters/criteria that narrow the search to uniquely identify a "product" record and match its price. We will make several queries to the `products` table each time analysing the results and adding new filters until we figure out a pattern that identifies the "FTP/SFTP/FTPS protocol enabled" component.

First of all, we need to get familiar with the table structure. Run `\d products` command to see the table's columns:

```sql
                 Table "public.products"
    Column     | Type  | Collation | Nullable | Default
---------------+-------+-----------+----------+----------
 productHash   | text  |           | not null |
 sku           | text  |           | not null |
 vendorName    | text  |           | not null |
 region        | text  |           |          |
 service       | text  |           | not null |
 productFamily | text  |           | not null | ''::text
 attributes    | jsonb |           | not null |
 prices        | jsonb |           | not null |
```

We can ignore `productHash` and `sku` columns here. The other columns are:

- `vendorName` defines a cloud provider. It can be `aws`/`azure`/`google`.
- `region` defines a cloud region. Different regions may have different pricing for the same services.
- `service` and `productFamily` define the service.
- `attributes` stores a JSON object with data about a specific feature in the service. It can have a different structure for different services and requires the most attention.
- `prices` stores the feature's single price or multiple prices (for example, if the service feature supports tiered pricing).

> Note: You may already notice how `ProductFilter` attributes map to this table.

Knowing that our cloud provider is AWS, we can introduce the first filter and see what information the table has.

Let's see what `service` and `productFamily` columns store:

```sql
SELECT DISTINCT "service", "productFamily" FROM products WHERE "vendorName" = 'aws';
```

```
           service            |                     productFamily
------------------------------+--------------------------------------------------------
 AmazonAppStream              | Stopped Instance
 AmazonS3                     | Storage
 AWSElementalMediaStore       | Fee
 AWSCostExplorer              | Storage
 ...
```

The query returns too many records. We could read through them all to find something resembling "AWS Transfer Family," but it is inefficient. Instead, we can significantly reduce the list by using another filter with a regular expression. Let's see what records contain the "transfer" word in their names (we can filter by `service` first, and if it's not enough, try `productFamily`):

```sql
SELECT DISTINCT "service", "productFamily" FROM products WHERE "vendorName" = 'aws' AND "service" ~* 'transfer';
```

```sql
     service     |     productFamily
-----------------+-----------------------
 AWSDataTransfer | DT-Data Transfer
 AWSDataTransfer | Data Transfer
 AWSTransfer     | AWS Transfer Family
 AWSTransfer     | AWS Transfer for SFTP
(4 rows)
```

Much better! Instead of hundreds of records, we reduced the search to four possibilities. But it is still unclear what exactly represents the service. The guess would be the last two rows, but it is uncertain. We need more information, and including the `attributes` column in the output may help. However, `DISTINCT` may not work anymore as `attributes` data is too diverse. To keep the list short, we will query for one region. The Terraform example's `us-east-1` will do:

```sql
SELECT DISTINCT "service", "productFamily", "attributes" FROM products WHERE "vendorName" = 'aws' AND "service" = 'AWSTransfer' AND region = 'us-east-1';
```

```
   service   |     productFamily     |                                                                                                                                        attributes
-------------+-----------------------+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 AWSTransfer | AWS Transfer Family   | {"data": "FTPS-None", "endpoint": "FTPS-Protocol-Hours", "location": "US East (N. Virginia)", "operation": "FTPS:S3", "usagetype": "USE1-ProtocolHours", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-Download-EFS", "endpoint": "SFTP-None-EFS", "location": "US East (N. Virginia)", "operation": "SFTP:EFS", "usagetype": "USE1-DownloadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-Download-Bytes", "endpoint": "SFTP-NONE", "location": "US East (N. Virginia)", "operation": "SFTP:S3", "usagetype": "USE1-DownloadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-Upload-EFS", "endpoint": "SFTP-None-EFS", "location": "US East (N. Virginia)", "operation": "SFTP:EFS", "usagetype": "USE1-UploadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer for SFTP | {"data": "SFTP-DATA", "endpoint": "SFTP-NONE", "location": "US East (N. Virginia)", "operation": "", "usagetype": "USE1-UPLOAD", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-NONE", "endpoint": "SFTP-Protocol-Hours", "location": "US East (N. Virginia)", "operation": "SFTP:S3", "usagetype": "USE1-ProtocolHours", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-Upload-Bytes", "endpoint": "SFTP-NONE", "location": "US East (N. Virginia)", "operation": "SFTP:S3", "usagetype": "USE1-UploadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer for SFTP | {"data": "SFTP-DATA", "endpoint": "SFTP-Endpoint", "location": "US East (N. Virginia)", "operation": "", "usagetype": "USE1-ENDPOINT", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "SFTP-None-EFS", "endpoint": "SFTP-Hours-EFS", "location": "US East (N. Virginia)", "operation": "SFTP:EFS", "usagetype": "USE1-ProtocolHours", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "FTP-None", "endpoint": "FTP-Protocol-Hours", "location": "US East (N. Virginia)", "operation": "FTP:S3", "usagetype": "USE1-ProtocolHours", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "FTPS-Download-Bytes", "endpoint": "FTPS-None", "location": "US East (N. Virginia)", "operation": "FTPS:S3", "usagetype": "USE1-DownloadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "FTP-Upload-Bytes", "endpoint": "FTP-None", "location": "US East (N. Virginia)", "operation": "FTP:S3", "usagetype": "USE1-UploadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "FTP-Download-Bytes", "endpoint": "FTP-None", "location": "US East (N. Virginia)", "operation": "FTP:S3", "usagetype": "USE1-DownloadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer for SFTP | {"data": "SFTP-DATA-DOWNLOADS", "endpoint": "SFTP-NONE", "location": "US East (N. Virginia)", "operation": "", "usagetype": "USE1-DOWNLOAD", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
 AWSTransfer | AWS Transfer Family   | {"data": "FTPS-Upload-Bytes", "endpoint": "FTPS-None", "location": "US East (N. Virginia)", "operation": "FTPS:S3", "usagetype": "USE1-UploadBytes", "regionCode": "us-east-1", "servicecode": "AWSTransfer", "servicename": "AWS Transfer Family", "locationType": "AWS Region"}
(15 rows)
```

Great! It looks like we found our service records. And we're down to only 15 records. They may have enough information to find the patterns to match a specific service feature.

We want to see what defines the "enabled protocol" and what values define the specific protocol. Among all `attributes`' values, we can highlight `data`, `operation` and `usagetype`. Let's see how they look without all other noise:

```sql
SELECT DISTINCT "service", "productFamily", "attributes"->>'data' AS data, "attributes"->>'operation' AS operation, "attributes"->>'usagetype' AS usagetype FROM products WHERE "vendorName" = 'aws' AND "service" = 'AWSTransfer' AND region = 'us-east-1';
```

```
   service   |     productFamily     |        data         | operation |     usagetype
-------------+-----------------------+---------------------+-----------+--------------------
 AWSTransfer | AWS Transfer Family   | FTP-Download-Bytes  | FTP:S3    | USE1-DownloadBytes
 AWSTransfer | AWS Transfer Family   | FTP-None            | FTP:S3    | USE1-ProtocolHours
 AWSTransfer | AWS Transfer Family   | FTP-Upload-Bytes    | FTP:S3    | USE1-UploadBytes
 AWSTransfer | AWS Transfer Family   | FTPS-Download-Bytes | FTPS:S3   | USE1-DownloadBytes
 AWSTransfer | AWS Transfer Family   | FTPS-None           | FTPS:S3   | USE1-ProtocolHours
 AWSTransfer | AWS Transfer Family   | FTPS-Upload-Bytes   | FTPS:S3   | USE1-UploadBytes
 AWSTransfer | AWS Transfer Family   | SFTP-Download-Bytes | SFTP:S3   | USE1-DownloadBytes
 AWSTransfer | AWS Transfer Family   | SFTP-Download-EFS   | SFTP:EFS  | USE1-DownloadBytes
 AWSTransfer | AWS Transfer Family   | SFTP-NONE           | SFTP:S3   | USE1-ProtocolHours
 AWSTransfer | AWS Transfer Family   | SFTP-None-EFS       | SFTP:EFS  | USE1-ProtocolHours
 AWSTransfer | AWS Transfer Family   | SFTP-Upload-Bytes   | SFTP:S3   | USE1-UploadBytes
 AWSTransfer | AWS Transfer Family   | SFTP-Upload-EFS     | SFTP:EFS  | USE1-UploadBytes
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA           |           | USE1-ENDPOINT
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA           |           | USE1-UPLOAD
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA-DOWNLOADS |           | USE1-DOWNLOAD
(15 rows)
```

Much cleaner! We can see different protocols and different "usage types". However, `AWS Transfer for SFTP` records stand out as there are no similar records for other protocols. It's time to bring the last piece of the puzzle - prices.

```sql
SELECT DISTINCT "service", "productFamily", "attributes"->>'data' AS data, "attributes"->>'operation' AS operation, "attributes"->>'usagetype' AS usagetype, prices FROM products WHERE "vendorName" = 'aws' AND "service" = 'AWSTransfer' AND region = 'us-east-1';
```

```
   service   |     productFamily     |        data         | operation |     usagetype      |                                                                                                                                                                                                               prices
-------------+-----------------------+---------------------+-----------+--------------------+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 AWSTransfer | AWS Transfer Family   | FTP-Download-Bytes  | FTP:S3    | USE1-DownloadBytes | {"3ee6535426a0bf75d16267a394ee64c0-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "3ee6535426a0bf75d16267a394ee64c0-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte downloaded over FTP from S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | FTP-None            | FTP:S3    | USE1-ProtocolHours | {"1e3ef5c6622041fcc4988e4258d5fc04-e7eda77c4cf52b2a5e814c7059c2e4c8": [{"USD": "0.3000000000", "unit": "Hourly", "priceHash": "1e3ef5c6622041fcc4988e4258d5fc04-e7eda77c4cf52b2a5e814c7059c2e4c8", "description": "$0.3 per Hour for FTP in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | FTP-Upload-Bytes    | FTP:S3    | USE1-UploadBytes   | {"8ed7d604e191d4a2905e674cd452e59c-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "8ed7d604e191d4a2905e674cd452e59c-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte uploaded over FTP to S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | FTPS-Download-Bytes | FTPS:S3   | USE1-DownloadBytes | {"87f1837c0ce7d43b79fc84e6dae1d9c1-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "87f1837c0ce7d43b79fc84e6dae1d9c1-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte downloaded over FTPS from S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | FTPS-None           | FTPS:S3   | USE1-ProtocolHours | {"c654a8df323f166e5b949e6ac2e0dcf0-e7eda77c4cf52b2a5e814c7059c2e4c8": [{"USD": "0.3000000000", "unit": "Hourly", "priceHash": "c654a8df323f166e5b949e6ac2e0dcf0-e7eda77c4cf52b2a5e814c7059c2e4c8", "description": "$0.3 per Hour for FTPS in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | FTPS-Upload-Bytes   | FTPS:S3   | USE1-UploadBytes   | {"7a18d0bdada1c2b960bf3448a3755295-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "7a18d0bdada1c2b960bf3448a3755295-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte uploaded over FTPS to S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-Download-Bytes | SFTP:S3   | USE1-DownloadBytes | {"b0baa77937b42384ffbb639cdb044509-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "b0baa77937b42384ffbb639cdb044509-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte downloaded over SFTP from S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-Download-EFS   | SFTP:EFS  | USE1-DownloadBytes | {"ccc12cb34ea30c096e7b822cee32afa6-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "ccc12cb34ea30c096e7b822cee32afa6-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte downloaded over SFTP from EFS in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-NONE           | SFTP:S3   | USE1-ProtocolHours | {"b7d8a0057f9d919c2a259ce626dad86a-e7eda77c4cf52b2a5e814c7059c2e4c8": [{"USD": "0.3000000000", "unit": "Hourly", "priceHash": "b7d8a0057f9d919c2a259ce626dad86a-e7eda77c4cf52b2a5e814c7059c2e4c8", "description": "$0.3 per Hour for SFTP in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-None-EFS       | SFTP:EFS  | USE1-ProtocolHours | {"95fa3c594bfda56481e4249bbb78b6a8-e7eda77c4cf52b2a5e814c7059c2e4c8": [{"USD": "0.3000000000", "unit": "Hourly", "priceHash": "95fa3c594bfda56481e4249bbb78b6a8-e7eda77c4cf52b2a5e814c7059c2e4c8", "description": "$0.3 per Hour for SFTP in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-Upload-Bytes   | SFTP:S3   | USE1-UploadBytes   | {"9fab128d84afd44ac5cbdf9c284715b8-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "9fab128d84afd44ac5cbdf9c284715b8-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte uploaded over SFTP to S3 in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer Family   | SFTP-Upload-EFS     | SFTP:EFS  | USE1-UploadBytes   | {"7bb4dfbd92f81e176c11a94420dd1730-dcaa14181f6c95f2f4f3e4ccf3fee63a": [{"USD": "0.0400000000", "unit": "GigaBytes", "priceHash": "7bb4dfbd92f81e176c11a94420dd1730-dcaa14181f6c95f2f4f3e4ccf3fee63a", "description": "$0.04 per GigaByte uploaded over SFTP to EFS in US East (N. Virginia)", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA           |           | USE1-ENDPOINT      | {"ad1cbaf5ec1686f7904c6f8a8ac2564a-e7eda77c4cf52b2a5e814c7059c2e4c8": [{"USD": "0.3000000000", "unit": "Hourly", "priceHash": "ad1cbaf5ec1686f7904c6f8a8ac2564a-e7eda77c4cf52b2a5e814c7059c2e4c8", "description": "$.30 per hour - SFTP Endpoint fee", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA           |           | USE1-UPLOAD        | {"333c11c5195600206cdc7a5ed5b83a57-b1ae3861dc57e2db217fa83a7420374f": [{"USD": "0.0400000000", "unit": "GB", "priceHash": "333c11c5195600206cdc7a5ed5b83a57-b1ae3861dc57e2db217fa83a7420374f", "description": "$0.04 per GB - SFTP Data Upload fee", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
 AWSTransfer | AWS Transfer for SFTP | SFTP-DATA-DOWNLOADS |           | USE1-DOWNLOAD      | {"e44e4cde7b4ef0fab2e019dd26038b24-b1ae3861dc57e2db217fa83a7420374f": [{"USD": "0.0400000000", "unit": "GB", "priceHash": "e44e4cde7b4ef0fab2e019dd26038b24-b1ae3861dc57e2db217fa83a7420374f", "description": "$0.04 per GB - SFTP Data Download fee", "endUsageAmount": "Inf", "purchaseOption": "on_demand", "startUsageAmount": "0", "effectiveDateStart": "2021-11-01T00:00:00Z"}]}
(15 rows)
```

Comparing and cross-matching our findings with [AWS Pricing page](https://aws.amazon.com/aws-transfer-family/pricing/) and [price calculator](https://calculator.aws/#/createCalculator/TransferFamily) can tell us the following things:

- We can ignore the `AWS Transfer for SFTP` product family because we can't apply that for other protocols resources.
- `attributes`' `operation` and `usagetype` can uniquely identify each variation of the service feature. We won't use `data` as its values are inconsistent. (Other cost components for downloaded/uploaded data can follow the same pattern).
- `EFS` storage type in `operation` column is missing for `FTP` and `FTPS`. However, AWS documentation says that these protocols support it.
- As the prices for `EFS` and `S3` are identical, we can use `operation` column with `<protocol>:S3` filter for both types.
- The other filter would use `usagetype` column and match `ProtocolHours`. However, its values start with `USE1-`. A quick check with a different region confirms that this is an acronym for the region, so we could use a regular expression to match it.
- The `prices` column contains only one price. The price filter should be simple.

> Note: Don't be discouraged if it takes several back and forths with the above queries. Cloud pricing is confusing, and it may take several tries to spot the correct pattern. It is also okay to adjust the filters later when adding more cost components.

It looks like we have all the necessary information to finish our first cost component.

Let's head back to the resource file and update the cost component's product filter:

```go
func (r *TransferServer) protocolEnabledCostComponent() *schema.CostComponent {
	// The pricing for all storage types is identical, but for some protocols
	// EFS prices are missing in the pricing API.
	storageType := "S3"

	protocol := "FTP"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s protocol enabled", protocol),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSTransfer"),
			ProductFamily: strPtr("AWS Transfer Family"),
			AttributeFilters: []*schema.AttributeFilter{
			{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-ProtocolHours$")},
			{Key: "operation", ValueRegex: regexPtr(fmt.Sprintf("^%s:%s$", protocol, storageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
```

Here's what we've changed:

- Updated the `ProductFilter` `Service` and `ProductFamily` values.
- Added two regular expression attribute filters. To make an explicit match, so we'll use `^...$` to match the beginning and the end of the string. `regexPtr` will take ensure to ignore the case.
- Added and used a `protocol` variable as we'll need to support several protocols.
- Added a `storageType` variable with a comment for some context. It may be a help if the resource needs updates in the future.

Let's rerun our tests!

```sh
ARGS="--run TestTransferServer -v" make test_aws
```

```
Output does not match golden file:

--- Expected
+++ Actual
@@ -1,5 +1,8 @@

- Name  Monthly Qty  Unit  Monthly Cost
-
- OVERALL TOTAL                   $0.00
+ Name                                 Monthly Qty  Unit   Monthly Cost
+
+ aws_transfer_server.transfer_server
+ └─ FTP protocol enabled                      730  hours       $219.00
+
+ OVERALL TOTAL                                                 $219.00
  ──────────────────────────────────
```

It shows the cost `$219.00` now, matching the price calculator page results! Success!

> Note: It is very important to ensure that there are **no warnings** about "Multiple products found", "No products found for" or "No prices found for" in the logs. These warnings indicate that the price filters have an issue.

But it shows only one protocol. What if the resource has several? Let's find out!

### Resource attributes mapping

According to [Terraform resource's documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/transfer_server#protocols) we have the following data:

- The attribute is an array. It can contain more than one value.
- The attribute is optional, meaning the resource may not have it defined.
- If the attribute is not present in the resource, the default value will be an array with a single value, `SFTP`.

Our cost component depends on this attribute. We are going to:

1. Define the attribute in the resource struct.
1. Map Terraform's attribute to the resource's one and pass its value to the resource build function.

The `TransferServer` struct at the top of the `internal/resources/aws/transfer_server.go` file defines the resource's content. That's where we are going to define the new attribute. It already has required attributes like `Address` (resource's name) and `Region`, let's add `Protocols` as an array of strings like so:

```go
type TransferServer struct {
	Address   string
	Region    string

	Protocols []string
}
```

> Note: Go is a statically typed programming language. That is why we need to define a type of the new struct attribute.

Next, open IaC provider's file corresponding to the resource `internal/providers/terraform/aws/transfer_server.go`:

```go
package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getTransferServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_transfer_server",
		RFunc: newTransferServer,
	}
}

func newTransferServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	r := &aws.TransferServer{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
```

The `newTransferServer` function is responsible for attributes mapping. `d` argument is a struct of `ResourceData` containing Terraform's data (in JSON format).

Update the function to look like this:

```go
func newTransferServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	protocols := []string{}

	if d.IsEmpty("protocols") {
		defaultProtocol := "SFTP"
		protocols = append(protocols, defaultProtocol)
	} else {
		for _, data := range d.Get("protocols").Array() {
			protocols = append(protocols, data.String())
		}
	}

	r := &aws.TransferServer{
		Address:   d.Address,
		Region:    region,
		Protocols: protocols,
	}

	...
}
```

Let's see what we do here:

- Declare an empty `protocols` array that will hold the resource's protocol values.
- As the attribute is optional `d` may or may not have it, we need to use its helper function `d.IsEmpty("protocols")` to check if its value is present.
- If it's missing, we add a default protocol, `SFTP`, per Terraform documentation.
- If the array is not empty, we extract its value and convert it to Go's array `d.Get("protocols").Array()` and then iterate over it to add its items to our `protocols` array. We also convert each item to Go's string type (`data.String()`).
- We add the resulting array to the new struct as `Protocols: protocols`.

Let's head back to resource file `internal/resources/aws/transfer_server.go` and
make use of our new attribute:

```go
func (r *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	for _, protocol := range r.Protocols {
		costComponents = append(costComponents, r.protocolEnabledCostComponent(protocol))
	}
	...
}

func (r *TransferServer) protocolEnabledCostComponent(protocol string) *schema.CostComponent {
	// The pricing for all storage types is identical, but for some protocols
	// EFS prices are missing in the pricing API.
	storageType := "S3"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s protocol enabled", protocol),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSTransfer"),
			ProductFamily: strPtr("AWS Transfer Family"),
			AttributeFilters: []*schema.AttributeFilter{
			{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-ProtocolHours$")},
			{Key: "operation", ValueRegex: regexPtr(fmt.Sprintf("^%s:%s$", protocol, storageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
```

The cost component now has a `protocol` string argument instead of a hardcoded "FTP" value. The `BuildResource` function iterates through the newly added `r.Protocols` list and creates cost components for each available protocol.

There is one last thing to do before we put all our work to the test: add a Terraform resource example with multiple protocols in `internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.tf`:

```
resource "aws_transfer_server" "transfer_server" {
  tags = {
    Name = "No protocols"
  }
}

resource "aws_transfer_server" "multiple_protocols" {
  protocols = ["SFTP", "FTPS", "FTP"]
  tags = {
    Name = "No protocols"
  }
}
```

Time to run the tests and check the results:

```sh
ARGS="--run TestTransferServer -v" make test_aws
```

```sh
Output does not match golden file:

--- Expected
+++ Actual
@@ -1,7 +1,15 @@

- Name  Monthly Qty  Unit  Monthly Cost
-
- OVERALL TOTAL                   $0.00
+ Name                                                    Monthly Qty  Unit   Monthly Cost
+
+ aws_transfer_server.multiple_protocols
+ ├─ FTP protocol enabled                                         730  hours       $219.00
+ ├─ FTPS protocol enabled                                        730  hours       $219.00
+ └─ SFTP protocol enabled                                        730  hours       $219.00
+
+ aws_transfer_server.transfer_server
+ └─ FTPS protocol enabled                                        730  hours       $219.00
+
+ OVERALL TOTAL                                                                    $876.00
  ──────────────────────────────────
-1 cloud resource was detected:
-∙ 1 was estimated, 1 includes usage-based costs, see https://infracost.io/usage-file
+2 cloud resources were detected:
+∙ 2 were estimated, 2 include usage-based costs, see https://infracost.io/usage-file
```

Congratulations! We added our first cost component! And we even test two cases: when protocols are not defined (and the default is used) and multiple protocols.

Let's celebrate this accomplishment by adding another cost component! It's time to see how usage-based cost components work.

### Add usage-based cost components

We distinguish the **price** of a resource from its **cost**. Price is the per-unit price advertised by a cloud vendor. The cost of a resource is calculated by multiplying its price by its usage. For example, an EC2 instance may have a price of $0.02 per hour, and if it's run for 100 hours (its usage), it'll cost $2.00. When adding resources to Infracost, we can always show their price, but if the resource has a usage-based cost component, we can't show its cost unless the user specifies the usage.

For the `aws_transfer_server` resource, the GitHub issue defines two such usage-based components: `Data uploaded` and `Data downloaded`.

Infracost supports passing usage data in through a usage YAML file. We'll add all necessary parts for the cost components first, and then we'll see how they use the YAML data.

Let's define new attributes in the resource struct:

```go
type TransferServer struct {
	Address   string
	Region    string
	Protocols []string

	// "usage" args
	MonthlyDataDownloadedGB *float64 `infracost_usage:"monthly_data_downloaded_gb"`
	MonthlyDataUploadedGB   *float64 `infracost_usage:"monthly_data_uploaded_gb"`
}
```

We expect the values to be floats. Here, we map the attributes to their counterparts in the usage file as `infracost_usage:"<attribute name>"`.

This way, the cost components can access these values.

Second, we need to update the resource's _usage schema_ to let the resource know _how_ to load the usage data. We'll add two new usage items. If CLI can't find the usage data, they will default to 0:

```go
// TransferServerUsageSchema defines a list of usage items for TransferServer.
var TransferServerUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_downloaded_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_uploaded_gb", DefaultValue: 0, ValueType: schema.Float64},
}
```

Next, let's add the cost components. From our price search step, we already know what the product filters will be:

- `usagetype` mentions `UploadBytes` and `DownloadBytes` so we can match it via regular expressions.
- The price is the same for any protocol and storage type so that we can pick one.

Here's how the components will look:

```go
func (r *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	...

	costComponents = append(costComponents, r.dataDownloadedCostComponent())
	costComponents = append(costComponents, r.dataUploadedCostComponent())

	...
}

func (r *TransferServer) dataDownloadedCostComponent() *schema.CostComponent {
	// The pricing is identical for all protocols and the traffic is combined
	operation := "FTP:S3"

	return &schema.CostComponent{
		Name:            "Data downloaded",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataDownloadedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSTransfer"),
			ProductFamily: strPtr("AWS Transfer Family"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-DownloadBytes$")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *TransferServer) dataUploadedCostComponent() *schema.CostComponent {
	// The pricing is identical for all protocols and the traffic is combined
	operation := "FTP:S3"

	return &schema.CostComponent{
		Name:            "Data uploaded",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataUploadedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSTransfer"),
			ProductFamily: strPtr("AWS Transfer Family"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-UploadBytes$")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
```

We've set up their unit to `GB` and passed the newly added `r.MonthlyDataDownloadedGB` and `r.MonthlyDataUploadedGB` to the `MonthlyQuantity` component's attribute.

> Note: As the YAML values are floats and `MonthlyQuantity` expects a decimal pointer, we have to make a type conversion. Infracost code has several helper functions like `floatPtrToDecimalPtr` to make it easier.

Let's run the test and see what's changed:

```sh
Output does not match golden file:

--- Expected
+++ Actual
@@ -1,7 +1,19 @@

- Name  Monthly Qty  Unit  Monthly Cost
-
- OVERALL TOTAL                   $0.00
+ Name                                                      Monthly Qty  Unit            Monthly Cost
+
+ aws_transfer_server.multiple_protocols
+ ├─ FTP protocol enabled                                           730  hours                $219.00
+ ├─ FTPS protocol enabled                                          730  hours                $219.00
+ ├─ SFTP protocol enabled                                          730  hours                $219.00
+ ├─ Data downloaded                                      Monthly cost depends on usage: $0.04 per GB
+ └─ Data uploaded                                        Monthly cost depends on usage: $0.04 per GB
+
+ aws_transfer_server.transfer_server
+ ├─ FTPS protocol enabled                                          730  hours                $219.00
+ ├─ Data downloaded                                      Monthly cost depends on usage: $0.04 per GB
+ └─ Data uploaded                                        Monthly cost depends on usage: $0.04 per GB
+
+ OVERALL TOTAL                                                                               $876.00
  ──────────────────────────────────
-1 cloud resource was detected:
-∙ 1 was estimated, 1 includes usage-based costs, see https://infracost.io/usage-file
+2 cloud resources were detected:
+∙ 2 were estimated, 2 include usage-based costs, see https://infracost.io/usage-file
```

Great! The new components pick up their prices and show up in the output.

Finally, we'll set the test usage parameters and verify that we calculate the cost correctly.

We will reuse the `aws_transfer_server.multiple_protocols` example resource from the Terraform file. It is good to add `_with_usage` to its name to highlight that it expects to have the usage parameters.

Open `internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.usage.yml` and add the following lines:

```yml
version: 0.1
resource_usage:
  aws_transfer_server.multiple_protocols_with_usage:
    monthly_data_downloaded_gb: 50
    monthly_data_uploaded_gb: 10
```

We must use the example's name in the YAML file to identify it during the CLI run correctly.

Rerunning the test should give us the following output:

```sh
ARGS="--run TestTransferServer -v" make test_aws
```

```sh
Output does not match golden file:

--- Expected
+++ Actual
@@ -1,7 +1,19 @@

- Name  Monthly Qty  Unit  Monthly Cost
-
- OVERALL TOTAL                   $0.00
+ Name                                                 Monthly Qty  Unit            Monthly Cost
+
+ aws_transfer_server.multiple_protocols_with_usage
+ ├─ FTP protocol enabled                                      730  hours                $219.00
+ ├─ FTPS protocol enabled                                     730  hours                $219.00
+ ├─ SFTP protocol enabled                                     730  hours                $219.00
+ ├─ Data downloaded                                            50  GB                     $2.00
+ └─ Data uploaded                                              10  GB                     $0.40
+
+ aws_transfer_server.transfer_server
+ ├─ FTPS protocol enabled                                     730  hours                $219.00
+ ├─ Data downloaded                                 Monthly cost depends on usage: $0.04 per GB
+ └─ Data uploaded                                   Monthly cost depends on usage: $0.04 per GB
+
+ OVERALL TOTAL                                                                          $878.40
  ──────────────────────────────────
-1 cloud resource was detected:
-∙ 1 was estimated, 1 includes usage-based costs, see https://infracost.io/usage-file
+2 cloud resources were detected:
+∙ 2 were estimated, 2 include usage-based costs, see https://infracost.io/usage-file
```

And now we have our usage cost. Perfect! After checking with the resource's pricing page and calculator, everything looks good. This is the output we want to get from the CLI. Let's finalize the golden file by running the tests with `-update` flag:

```sh
ARGS="--run TestTransferServer -v -update" make test_aws
```

```sh
=== RUN   TestTransferServerGoldenFile
=== PAUSE TestTransferServerGoldenFile
=== CONT  TestTransferServerGoldenFile
--- PASS: TestTransferServerGoldenFile (2.77s)
PASS
ok      github.com/infracost/infracost/internal/providers/terraform/aws 4.330s
```

It passes. If we look at `internal/providers/terraform/aws/testdata/transfer_server_test/transfer_server_test.golden` it'll have the result output.

The last thing is to update `infracost-usage-example.yml` with an example of usage parameters so Infracost users can see how they can use them. Open the file and add the following block:

```yml
  aws_transfer_server.my_transfer_server:
    monthly_data_downloaded_gb: 50 # Monthly data downloaded over enabled protocols in GB.
    monthly_data_uploaded_gb: 10   # Monthly data uploaded over enabled protocols in GB.
```

The Infracost convention is to place it in alphabetical order for easier navigation, use `my_<resource_name>` as the resource's name, and provide a comment with parameter description.

We are at the finish line! We'll run `make fmt` and `make lint` commands to verify that our code is nice and tidy.

Congratulations! We've added all the required cost components! Let's wrap things up and prepare a pull request!

## Add free resources

If a GitHub issue mentions free resources related to the one with price (or you notice them yourself), we need to add them to a free resources list. This way, if CLI detects them, it won't mark them as "not supported," which reduces output noise.

There are three free resources related to `aws_transfer_server` resource. Let's open `internal/providers/terraform/aws/registry.go` file and find `FreeResources` array. We'll add our free resources into it, finding the right place according to the alphabetical order:

```go
	// AWS Transfer Family
	"aws_transfer_access",
	"aws_transfer_ssh_key",
	"aws_transfer_user",
```

## Update Infracost docs

[Infracost documentation](https://www.infracost.io/docs/supported_resources/overview/) lists all supported paid and free resources. As we add a new resource, let's also update the docs.

1. Create a fork of [docs repo](https://github.com/infracost/docs).
1. Create a new branch and update [this file](https://github.com/infracost/docs/blob/master/docs/supported_resources/aws.md):
  - Add your paid resource in the `Paid resources` section.
  - If you added free resources, include them in the `Free resources` section.

  > Note: we keep an alphabetical order in these tables for easier navigation.

1. Commit your changes with the `docs:` prefix, for example, `docs: add AWS Transfer Family resources` message. Push it to GitHub and create a pull request to Infracost's docs repository.

## Prepare a Pull Request for review

Last step is to commit our changes and create a pull request in [Infracost repo](https://github.com/infracost/infracost).

1. Commit your changes using `feat(aws):` prefix (Infracost team uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)).
1. Push the branch to your fork and create a new pull request in the Infracost repo using the following template:

```markdown
## Objective:

Add support for <RESOURCE NAME>. Fixes #<ISSUE NUMBER>.

## Pricing details:

TODO: Add all the details you find essential for the pull request.

## Status:

- [ ] Generated the resource files
- [ ] Updated the internal/resources file
- [ ] Updated the internal/provider/terraform/.../resources file
- [ ] Added usage parameters to infracost-usage-example.yml
- [ ] Added test cases without usage-file
- [ ] Added test cases with usage-file
- [ ] Compared test case output to cloud cost calculator
- [ ] Created a PR to update "Supported Resources" in the [docs](<LINK TO DOCS PULL REQUEST>)

## Issues:

None
```

Mark all the checkboxes you've done and wait for somebody from the Infracost team to review your changes.

🎉  Congratulations! You've added your first Infracost resource. Great job, thank you! Every new resource makes Infracost more useful and gets users closer to loving their cloud bill! 💰📉

<p align="center">
  <img src="https://infracost-public-dumps.s3.amazonaws.com/images/congratulations.gif" alt="Congratulations!" />
</p>

---

## Appendix

### Resource edge cases

The following edge cases can be handled in the resource files:

- free resources: if there are certain conditions that can be checked inside a resource Go file, which means there are no cost components for the resource, return a `NoPrice: true` and `IsSkipped: true` response as shown below.

  ```go
	// Gateway endpoints don't have a cost associated with them
	if vpcEndpointType == "Gateway" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}
  ```

- unsupported resources: if there are certain conditions that can be checked inside a resource Go file, which means that the resource is not yet supported, log a warning to explain what is not supported and return a `nil` response as shown below.

  ```go
	if d.Get("placement_tenancy").String() == "host" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	}
  ```

- use `IgnoreIfMissingPrice: true` if you need to lookup a price in the Cloud Pricing API and NOT add it if there is no price. We use it for EBS Optimized instances since we don't know if they should have that cost component without looking it up.

### References to other resources

If you need access to other resources referenced by the resource you're adding, you can specify `ReferenceAttributes`. The following example uses this because the price for `aws_ebs_snapshot` depends on the size of the referenced volume. You should always check the array length returned by `d.References` to avoid panics. You can also do nested lookups, e.g. `ReferenceAttributes: []string{"alias.0.name"}` to lookup the name of the first alias.

  ```go
	func GetEBSSnapshotRegistryItem() *schema.RegistryItem {
		return &schema.RegistryItem{
			Name:                "aws_ebs_snapshot",
			RFunc:               NewEBSSnapshot,
			ReferenceAttributes: []string{"volume_id"}, // Load the reference
		}
	}

	func NewEBSSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
		volumeRefs := d.References("volume_id") // Get the reference

		// Always check the array length to avoid panics as `d.References` might not find the reference, e.g. it might point to another module via a data resource.
		if len(volumeRefs) > 0 {
			gbVal = decimal.NewFromFloat(volumeRefs[0].Get("size").Float())
		}
		// ...
	}
  ```

#### Reverse references

Sometimes you need access to a resource that refers to the resource you're adding. In this case you can add `ReferenceAttributes` to both resources. The resource with the reference field should be added as normal (`reference_field`) and resource that is the target of the reference should add a type qualified reference to the same field (`referring_type.reference_field`). For example, if an `aws_ebs_volume` needed the list of snapshots that reference it, the `aws_ebs_snapshot` would be the same as above and the volume would be:

```go
	func GetEBSVolumeRegistryItem() *schema.RegistryItem {
		return &schema.RegistryItem{
			Name:                "aws_ebs_volume",
			RFunc:               NewEBSVolume,
			// This only works if aws_ebs_snapshot has defined "volume_id" as a ReferenceAttribute
			ReferenceAttributes: []string{"aws_ebs_snapshot.volume_id"},
		}
	}

	func NewEBSVolume(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
		snapshotReverseRefs := d.References("aws_ebs_snapshot.volume_id") // Get the reference
		// ...
	}
  ```

#### Custom reference ids

By default, references are matched using an AWS ARN or the id field (`d.Get("id")`). Sometimes cloud providers use name or another field for references. In this case you can provide a 'custom reference id function' that generates additional ids used to match references. In this example, a custom id is needed because `aws_appautoscaling_target.resource_id` contains a `table/<name>` string when referencing an `aws_dynamodb_table`.

```go
	func getAppAutoscalingTargetRegistryItem() *schema.RegistryItem {
		return &schema.RegistryItem{
			Name:                "aws_appautoscaling_target",
			RFunc:               NewAppAutoscalingTargetResource,
			ReferenceAttributes: []string{"resource_id"},
		}
	}

	func NewAppAutoscalingTargetResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
		// ...
	}
  ```

```go
	func getDynamoDBTableRegistryItem() *schema.RegistryItem {
		return &schema.RegistryItem{
			Name:            "aws_dynamodb_table",
			RFunc:           NewDynamoDBTableResource,
			CustomRefIDFunc: func (d *schema.ResourceData) []string {
				// returns an id that will match the custom format used by aws_appautoscaling_target.resource_id
				name := d.Get("name").String()
				if name != "" {
					return []string{"table/" + name}
				}
				return nil
			},
		}
	}

	func NewAppAutoscalingTargetResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
		// ...
	}
```

### Google zone mappings

If the resource has a `zone` key, if they have a zone key, use this logic to get the region:

  ```go
	region := d.Get("region").String()
	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}
  ```

### Azure zone mappings

Unless the resource has global or zone-based pricing, you should add a custom GetRegion function to the RegistryItem. This should use the `lookupRegion` function that has the following signature: `lookupRegion(d *schema.ResourceData, parentResourceKeys []string) string`. The second parameter for this function is an optional list of parent resources where the region can be found. See the following examples of how this method is used in other Azure resources.

  ```go
	func getAppServiceCertificateBindingRegistryItem() *schema.RegistryItem {
		return &schema.RegistryItem{
			Name:  "azurerm_app_service_certificate_binding",
			RFunc: NewAzureRMAppServiceCertificateBinding,
			ReferenceAttributes: []string{
				"certificate_id",
				"resource_group_name",
			},
      GetRegion: func(d *schema.ResourceData) string {
        return lookupRegion(d, []string{"certificate_id", "resource_group_name"})
      },
		}
	}
  ```

### Region usage

A number of resources have usage costs which vary on a per-region basis. This means that you often have to define a usage file with a complex map key with all the cloud provider regions. For example, `google_artifact_registry_repository` can have any number of Google regions under the `monthly_egress_data_transfer_gb` usage parameter:

```yaml
  google_artifact_registry_repository.artifact_registry:
    storage_gb: 150
    monthly_egress_data_transfer_gb:
      us_east1: 100
      us_west1: 100
      australia_southeast1: 100
      europe_north1: 100
      southamerica_east1: 100
```

If you have a resource like this, rather than defining your own usage field, you should use one of the shared `RegionsUsage` structs that handle this structure for you. These structs can be found in both [the `google`](../internal/resources/google/util.go) & [`aws`](../internal/resources/aws/util.go) resource util files.

These can simply be embedded into a struct field like so:

```go
type MyResource struct {
	// ...
	MonthlyDataProcessedGB *RegionsUsage `infracost_usage:"monthly_processed_gb"`
}
```

And then after `PopulateUsage` is called it can be accessed to retrieve set values. `RegionsUsage` helper struct also comes with a `Values` method that returns the set values as a `slice` with key/value pairs that is helpful to iterate over to create cost components, e.g with a usage like so:

```yaml
  my_resource.resource:
    monthly_processed_gb:
      europe_north1: 100
      southamerica_east1: 200
```

running:

```go
func (r *MyResource) BuildResource() *schema.Resource {
	values := r.MonthlyDataProcessedGB.Values()

	for _, v := range values {
		fmt.Println("%s => %2.f", v.Key, v.Value)
	}

	// ...
}
```

would print:


```bash
europe-north1 => 100.00
southamerica-east1 => 200.00
```

#### Adding new regions

Every so often cloud providers add locations/regions to their cloud infrastructure. When this happens we need to update shared provider variables so that the new regions are available in usage files.

#### AWS

Common usage structs are defined in the [`internal/resources/aws/util.go`](../internal/resources/aws/util.go) file. You'll need to update:

```go
var RegionMapping = map[string]string{
	"us-gov-west-1":   "AWS GovCloud (US-West)",
	// Add the new region here with the aws code mapping to the region name
	// as defined here: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
}
```

```go
type RegionsUsage struct {
	USGovWest1   *float64 `infracost_usage:"us_gov_west_1"`
	// Add your new region here with the infracost_usage struct tag
	// representing an underscored version of the aws region code
	// e.g: eu-west-1 => eu_west_1.
	//
	// The struct field type must be *float
}
```

```go
var RegionUsageSchema = []*schema.UsageItem{
	{Key: "us_gov_west_1", DefaultValue: 0, ValueType: schema.Float64},
	// Finally, add your new region to the usage schema.
	// Set the Key as the underscored code of the region (as outlined
	// in the prior RegionsUsage struct). Then set DefaultValue as 0
	// and the ValueType to schema.Float64.
}
```

#### Google

Common usage structs are defined in the [`internal/resources/google/util.go`](../internal/resources/google/util.go) file. You'll need to update:

```go
type RegionsUsage struct {
	AsiaEast1              *float64 `infracost_usage:"asia_east1"`
	// Add your new region here with the infracost_usage struct tag
	// representing an underscored version of the google location code
	// e.g: eu-west-1 => eu_west_1.
	//
	// The struct field type must be *float
}
```

```go
var RegionUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_east1"},
	// Finally, add your new region to the usage schema.
	// Set the Key as the underscored code of the location (as outlined
	// in the prior RegionsUsage struct). Then set DefaultValue as 0
	// and the ValueType to schema.Float64.
}
```
