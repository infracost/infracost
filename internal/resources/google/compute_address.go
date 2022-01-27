package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type ComputeAddress struct {
	Address     string
	Region      string
	AddressType string
}

var ComputeAddressUsageSchema = []*schema.UsageItem{}

func (r *ComputeAddress) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeAddress) BuildResource() *schema.Resource {
	region := r.Region

	addressType := r.AddressType
	if strings.ToLower(addressType) == "internal" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: ComputeAddressUsageSchema,
		}
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			standardVMComputeAddress(),
			preemptibleVMComputeAddress(),
			unusedVMComputeAddress(region),
		}, UsageSchema: ComputeAddressUsageSchema,
	}
}

func standardVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if used by standard VM)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Standard VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("744"),
		},
	}
}

func preemptibleVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if used by preemptible VM)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Spot Preemptible VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

func unusedVMComputeAddress(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if unused)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Static Ip Charge")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}
