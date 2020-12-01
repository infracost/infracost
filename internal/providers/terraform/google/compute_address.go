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

func NewComputeAddress(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}
	addressType := d.Get("address_type").String()
	if addressType != "EXTERNAL" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			standardVMComputeAddress(region),
			preemptibleVMComputeAddress(region),
			unusedVMComputeAddress(region),
		},
	}
}

func standardVMComputeAddress(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Static and ephemeral IP addresses in use on standard VM instances",
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
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}

func preemptibleVMComputeAddress(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Static and ephemeral IP addresses in use on preemptible VM instances",
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
		Name:           "Static IP address (assigned but unused)",
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
