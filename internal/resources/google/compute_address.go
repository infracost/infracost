package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type ComputeAddress struct {
	Address                string
	Region                 string
	AddressType            string
	Purpose                string
	InstancePurchaseOption string
}

func (r *ComputeAddress) CoreType() string {
	return "ComputeAddress"
}

func (r *ComputeAddress) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ComputeAddress) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeAddress) BuildResource() *schema.Resource {
	addressType := r.AddressType
	isFreePurpose := r.Purpose != "" && strings.ToLower(r.Purpose) != "gce_endpoint"

	if strings.ToLower(addressType) == "internal" || isFreePurpose {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	costComponents := []*schema.CostComponent{}

	switch r.InstancePurchaseOption {
	case "on_demand":
		costComponents = append(costComponents, r.standardVMComputeAddress())
	case "preemptible":
		costComponents = append(costComponents, r.preemptibleVMComputeAddress())
	default:
		costComponents = append(costComponents, r.unusedVMComputeAddress())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ComputeAddress) standardVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (standard VM)",
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
			StartUsageAmount: strPtr("696"),
		},
	}
}

func (r *ComputeAddress) preemptibleVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (preemptible VM)",
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

func (r *ComputeAddress) unusedVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (unused)",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("^Static Ip Charge.*")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}
