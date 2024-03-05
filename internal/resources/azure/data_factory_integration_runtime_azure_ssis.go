package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// DataFactoryIntegrationRuntimeAzureSSIS struct represents Data Factory's
// Azure-SSIS runtime.
//
// Resource information: https://azure.microsoft.com/en-us/services/data-factory/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/data-factory/ssis/
type DataFactoryIntegrationRuntimeAzureSSIS struct {
	Address string
	Region  string

	Instances       int64
	InstanceType    string
	Enterprise      bool
	LicenseIncluded bool
}

func (r *DataFactoryIntegrationRuntimeAzureSSIS) CoreType() string {
	return "DataFactoryIntegrationRuntimeAzureSSIS"
}

func (r *DataFactoryIntegrationRuntimeAzureSSIS) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the DataFactoryIntegrationRuntimeAzureSSIS.
// It uses the `infracost_usage` struct tags to populate data into the DataFactoryIntegrationRuntimeAzureSSIS.
func (r *DataFactoryIntegrationRuntimeAzureSSIS) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid DataFactoryIntegrationRuntimeAzureSSIS struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *DataFactoryIntegrationRuntimeAzureSSIS) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.computeCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// computeCostComponent returns a cost component for cluster configuration.
func (r *DataFactoryIntegrationRuntimeAzureSSIS) computeCostComponent() *schema.CostComponent {
	tier := "Standard"
	if r.Enterprise {
		tier = "Enterprise"
	}

	license := "License Included"
	licenseTitle := ", license included"
	if !r.LicenseIncluded {
		license = "AHB"
		licenseTitle = ""
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Compute (%s, %s%s)", r.InstanceType, tier, licenseTitle),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.Instances)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Data Factory v2"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr(license)},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", r.InstanceType))},
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("^SSIS %s", tier))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
