package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
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
		Name:           d.Address,
		CostComponents: []*schema.CostComponent{computeAddress(region)},
	}
}

func computeAddress(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Static and ephemeral IP addresses in use on standard VM instances",
		Unit:           "hours",
		UnitMultiplier: 1,
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
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}
