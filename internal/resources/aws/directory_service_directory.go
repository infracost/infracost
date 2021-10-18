package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

const (
	dtMicrosoftAD = "Microsoft AD"
)

var (
	directoryServiceDirectorySchema = []*schema.UsageItem{
		{
			Key: "monthly_data_processed_gb",
			DefaultValue: &usage.ResourceUsage{
				Name:  "monthly_data_processed_gb",
				Items: RegionUsageSchema,
			},
			ValueType: schema.SubResourceUsage,
		},
		{
			Key:          "shared_accounts",
			DefaultValue: 0,
			ValueType:    schema.Int64,
		},
	}

	awsVendorFilter       = strPtr("aws")
	directorySvcSvcFilter = strPtr("AWSDirectoryService")
	productFmlyFilter     = strPtr("AWS Directory Service")
)

// DirectoryServiceDirectory represents a single AWS Directory Service "Directory".
// AWS Directory Service has three main types: Microsoft AD, AD Connector & Simple AD.
// All the types are meant to support some form of Active Directory integration.
// Microsoft Active Directory is used by Windows applications to manage access and enable single sign-on.
// For example, you can manage access to Microsoft SharePoint using different Microsoft Active Directory security groups.
//
// The Microsoft AD type is a fully managed Microsoft Active Directory service.
// The Simple AD type is a fully managed Samba service which is compatible with Microsoft Active Directory.
// The AD Connector type is a gateway with which you can redirect directory requests to your on-premises Microsoft Active Directory.
//
// Read more about Directory service here: https://aws.amazon.com/directoryservice/
// Microsoft Active Directory here: https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_microsoft_ad.html
// Other Supported Active Directory types here:
//		https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_simple_ad.html
//		https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_simple_ad.html
//
// DirectoryServicePricing pricing is based on
//
// > Hourly price based on the type and edition (only Microsoft AD) of the directory service directory
// > Additional hourly price added directory per account/vpc the directory is shared with (only Microsoft AD)
// > Costs for data transfer out (on a per-region basis)
//
// More information on pricing can be found here:
//		https://aws.amazon.com/directoryservice/pricing/
// 		https://aws.amazon.com/directoryservice/other-directories-pricing/
type DirectoryServiceDirectory struct {
	// Address is the unique name of the resource in terraform/cloudfront.
	Address string
	// Region is the aws region the DirectoryServiceDirectory is provisioned within
	Region string
	// RegionName is the full region name used in product filters for the DirectoryService
	RegionName string

	// Type is the directory type. It can be one of (SimpleAD|ADConnector|MicrosoftAD)
	Type string
	// Edition is the edition of the MicrosoftAD type directory service. This field
	// is only applicable with MicrosoftAD and can either be (Standard|Enterprise).
	Edition string
	// The size of the directory, only applicable if the type is SimpleAD or ADConnector.
	// Values can be either (Small|Large)
	Size string

	// MonthlyDataProcessedDB represents data transfer from the directory to given
	// regions. This field is built from usage cost file and requires users to specify
	// estimates for each region they wish to see.
	MonthlyDataProcessedGB *RegionsUsage `infracost_usage:"monthly_data_processed_gb"`

	// SharedAccounts represents the number of accounts/vpcs the directory is shared with.
	// This cost is only applicable if the type of directory is MicrosoftAD.
	// Directory Service sharing support is not supported by terraform aws at this time.
	// Therefore, this field is built from the usage cost file. An open issue referencing
	// shared directory support here: https://github.com/hashicorp/terraform-provider-aws/issues/6003
	SharedAccounts *int64 `infracost_usage:"shared_accounts"`
}

// PopulateUsage parses the u schema.Usage into the DirectoryServiceDirectory.
// It uses the `infracost_usage` struct tags to populate data into the DirectoryServiceDirectory.
// monthly_data_transfer_out_gb.
func (d *DirectoryServiceDirectory) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(d, u)
}

// BuildResource builds a schema.Resource from a valid DirectoryServiceDirectory.
func (d *DirectoryServiceDirectory) BuildResource() *schema.Resource {
	// set the size based on the type of resource.
	// MicrosoftAD uses edition in product filters under the directorySize attribute.
	size := d.Size
	if d.Type == dtMicrosoftAD {
		size = d.Edition
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Directory service (%s, %s)", d.Type, size),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:       awsVendorFilter,
				Region:           strPtr(d.Region),
				Service:          directorySvcSvcFilter,
				ProductFamily:    productFmlyFilter,
				AttributeFilters: d.getAttributeFiltersForDirectory(size),
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if d.MonthlyDataProcessedGB != nil {
		for _, r := range d.MonthlyDataProcessedGB.Values() {
			// no outbound data transfer from the same region
			if r.Key == d.Region {
				continue
			}

			toLocation, ok := regionMapping[r.Key]
			if !ok {
				log.Warnf("Skipping resource %s usage cost: Outbound data transfer. Could not find mapping for region %s", d.Address, r.Key)
				continue
			}

			costComponents = append(costComponents, &schema.CostComponent{
				Name:            fmt.Sprintf("Outbound data transfer (from %s, to %s)", d.Region, r.Key),
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.Value)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    awsVendorFilter,
					Service:       strPtr("AWSDataTransfer"),
					ProductFamily: strPtr("Data Transfer"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "transferType", Value: strPtr("InterRegion Outbound")},
						{Key: "fromLocation", Value: strPtr(d.RegionName)},
						{Key: "toLocation", Value: strPtr(toLocation)},
					},
				},
			})
		}

		if d.SharedAccounts != nil && d.Type == dtMicrosoftAD {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Directory sharing",
				Unit:            "accounts",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(*d.SharedAccounts)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    awsVendorFilter,
					Region:        strPtr(d.Region),
					Service:       directorySvcSvcFilter,
					ProductFamily: productFmlyFilter,
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "directorySize", Value: strPtr(size)},
						{Key: "directoryType", Value: strPtr("Shared " + d.Type)},
						{Key: "location", Value: strPtr(d.RegionName)},
					},
				},
			})
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		UsageSchema:    directoryServiceDirectorySchema,
	}
}

func (d DirectoryServiceDirectory) getAttributeFiltersForDirectory(size string) []*schema.AttributeFilter {
	if d.Type == dtMicrosoftAD {
		return []*schema.AttributeFilter{
			{Key: "directorySize", Value: strPtr(size)},
			{Key: "directoryType", Value: strPtr(d.Type)},
			{Key: "location", Value: strPtr(d.RegionName)},
		}
	}

	// Simple AD and AD Connector types have directoryType fields of "Shared AD or AD Connector"
	// depending on the size. Therefore, we'll build a regex to match one of the names.
	return []*schema.AttributeFilter{
		{Key: "directorySize", Value: strPtr(size)},
		{Key: "directoryType", ValueRegex: strPtr(fmt.Sprintf(`/%s/i`, strings.ReplaceAll(d.Type, " ", `\s`)))},
		{Key: "location", Value: strPtr(d.RegionName)},
	}
}
