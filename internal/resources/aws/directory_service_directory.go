package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

const (
	dtMicrosoftAD = "Microsoft AD"
)

var (
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
//
//	https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_simple_ad.html
//	https://docs.aws.amazon.com/directoryservice/latest/admin-guide/directory_simple_ad.html
//
// # DirectoryServicePricing pricing is based on
//
// > Hourly price based on the type and edition (only Microsoft AD) of the directory service directory
// > Additional hourly price added directory per account/vpc the directory is shared with (only Microsoft AD)
// > Costs for data transfer out (on a per-region basis)
//
// More information on pricing can be found here:
//
//	https://aws.amazon.com/directoryservice/pricing/
//	https://aws.amazon.com/directoryservice/other-directories-pricing/
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

	// AdditionalDomainControllers represents a usage cost definition for the number controllers
	// above the default value (2) that are provisioned in this directory service.
	AdditionalDomainControllers *float64 `infracost_usage:"additional_domain_controllers"`

	// SharedAccounts represents the number of accounts/vpcs the directory is shared with.
	// This cost is only applicable if the type of directory is MicrosoftAD.
	// Directory Service sharing support is not supported by terraform aws at this time.
	// Therefore, this field is built from the usage cost file. An open issue referencing
	// shared directory support here: https://github.com/hashicorp/terraform-provider-aws/issues/6003
	SharedAccounts *float64 `infracost_usage:"shared_accounts"`
}

func (d *DirectoryServiceDirectory) CoreType() string {
	return "DirectoryServiceDirectory"
}

func (d *DirectoryServiceDirectory) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "additional_domain_controllers", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "shared_accounts", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.Usage into the DirectoryServiceDirectory.
// It uses the `infracost_usage` struct tags to populate data into the DirectoryServiceDirectory.
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
		d.domainControllerCostComponent(
			2, // directory service provisions a minimum of 2 controllers
			fmt.Sprintf("Directory service (%s, %s)", d.Type, size),
			size,
		),
	}

	if d.AdditionalDomainControllers != nil {
		costComponents = append(
			costComponents,
			d.additionalDomainControllerCostComponent(*d.AdditionalDomainControllers, size),
		)
	}

	if d.SharedAccounts != nil && d.Type == dtMicrosoftAD {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Directory sharing",
			Unit:           "accounts",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromFloat(*d.SharedAccounts)),
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
			UsageBased: true,
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		UsageSchema:    d.UsageSchema(),
	}
}

func (d DirectoryServiceDirectory) domainControllerCostComponent(amount float64, name, size string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "controllers",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromFloat(amount)),
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
	}
}

func (d DirectoryServiceDirectory) additionalDomainControllerCostComponent(amount float64, size string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Additional domain controllers",
		Unit:           "controllers",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromFloat(amount)),
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
		UsageBased: true,
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
