package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetComputeAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_address",
		RFunc:               NewComputeAddress,
		ReferenceAttributes: []string{},
	}
}

func GetComputeGlobalAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_global_address",
		RFunc:               NewComputeAddress,
		ReferenceAttributes: []string{},
	}
}

func NewComputeAddress(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	addressType := d.Get("address_type").String()
	if addressType == "INTERNAL" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			standardVMComputeAddress(),
			preemptibleVMComputeAddress(),
			unusedVMComputeAddress(region),
		},
	}
}

func standardVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if used by standard VM)",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: nil,
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
			StartUsageAmount: strPtr("744"), // use the non-free tier
		},
	}
}

func preemptibleVMComputeAddress() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if used by preemptible VM)",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: nil,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Preemptible VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}

func unusedVMComputeAddress(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if unused)",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: nil,
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
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}
