package google

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// SecretManagerSecretVersion represents one Google Secret Manager Secret's Version resource.
//
// The cost of active secret version depends on the number of replication
// locations specified by its parent secret. If it's more than one then the price
// is multiplied by the locations' quantity.
// Pricing API includes Free Tier, but it's not used.
//
// More resource information here: https://cloud.google.com/secret-manager
// Pricing information here: https://cloud.google.com/secret-manager/pricing
type SecretManagerSecretVersion struct {
	Address              string
	Region               string
	ReplicationLocations int64

	// "usage" args
	MonthlyAccessOperations *int64 `infracost_usage:"monthly_access_operations"`
}

func (r *SecretManagerSecretVersion) CoreType() string {
	return "SecretManagerSecretVersion"
}

func (r *SecretManagerSecretVersion) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_access_operations", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the SecretManagerSecretVersion.
// It uses the `infracost_usage` struct tags to populate data into the SecretManagerSecretVersion.
func (r *SecretManagerSecretVersion) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SecretManagerSecretVersion.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SecretManagerSecretVersion) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.activeSecretVersionsCostComponents()...)
	costComponents = append(costComponents, r.accessOperationsCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// activeSecretVersionsCostComponents returns a cost component for the Active Secret
// Version. By default it represents one version.
// The cost is multiplied by the number of replication locations. Free tier
// pricing is excluded.
func (r *SecretManagerSecretVersion) activeSecretVersionsCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Active secret versions",
			Unit:            "versions",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: intPtrToDecimalPtr(&r.ReplicationLocations),
			ProductFilter:   r.buildProductFilter("Secret version replica storage"),
			PriceFilter:     r.buildPriceFilter("6"),
		},
	}
}

// accessOperationsCostComponents returns a cost component for Secret Version's Access
// Operations. Free tier pricing is excluded.
func (r *SecretManagerSecretVersion) accessOperationsCostComponents() []*schema.CostComponent {
	multiplier := 10000

	return []*schema.CostComponent{
		{
			Name:            "Access operations",
			Unit:            "10K requests",
			UnitMultiplier:  decimal.NewFromInt(int64(multiplier)),
			MonthlyQuantity: intPtrToDecimalPtr(r.MonthlyAccessOperations),
			ProductFilter:   r.buildProductFilter("Secret access operations"),
			PriceFilter:     r.buildPriceFilter(fmt.Sprint(multiplier)),
			UsageBased:      true,
		},
	}
}

// buildProductFilter creates a product filter for Secret Manager's Secret
// product.
func (r *SecretManagerSecretVersion) buildProductFilter(description string) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:    strPtr("gcp"),
		Region:        strPtr(r.Region),
		Service:       strPtr("Secret Manager"),
		ProductFamily: strPtr("ApplicationServices"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", description))},
		},
	}
}

// buildPriceFilter creates a price filter based on start usage amount to ignore
// free tier pricing.
func (r *SecretManagerSecretVersion) buildPriceFilter(startUsageAmount string) *schema.PriceFilter {
	return &schema.PriceFilter{
		PurchaseOption:   strPtr("OnDemand"),
		StartUsageAmount: strPtr(startUsageAmount),
	}
}
