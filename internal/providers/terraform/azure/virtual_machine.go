package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// Parse from instance type value to Azure SKU name.
func parseVMSKUName(instanceType string) string {
	s := strings.ReplaceAll(instanceType, "Standard_", "")
	s = strings.ReplaceAll(s, "Basic_", "")
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

func ultraSSDReservationCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Ultra disk reservation (if unattached)",
		Unit:           "vCPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: nil,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Ultra Disks")},
				{Key: "skuName", Value: strPtr("Ultra LRS")},
				{Key: "meterName", Value: strPtr("Reservation per vCPU Provisioned")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func osDiskSubResource(region string, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.Get("os_disk").Array()) == 0 {
		return nil
	}

	diskData := d.Get("os_disk").Array()[0]

	var monthlyDiskOperations *decimal.Decimal

	if u != nil && u.Get("os_disk.monthly_disk_operations").Exists() {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(u.Get("os_disk.monthly_disk_operations").Int()))
	}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: managedDiskCostComponents(region, diskData, monthlyDiskOperations),
	}
}
