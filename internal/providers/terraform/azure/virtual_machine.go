package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var diskSizeMap = map[string][]struct {
	Name string
	Size int
}{
	"Standard_LRS": {
		{"S4", 32},
		{"S6", 64},
		{"S10", 128},
		{"S15", 256},
		{"S20", 512},
		{"S30", 1024},
		{"S40", 2048},
		{"S50", 4096},
		{"S60", 8192},
		{"S70", 16384},
		{"S80", 32767},
	},
	"StandardSSD_LRS": {
		{"E1", 4},
		{"E2", 8},
		{"E3", 16},
		{"E4", 32},
		{"E6", 64},
		{"E10", 128},
		{"E15", 256},
		{"E20", 512},
		{"E30", 1024},
		{"E40", 2048},
		{"E50", 4096},
		{"E60", 8192},
		{"E70", 16384},
		{"E80", 32767},
	},
	"Premium_LRS": {
		{"P1", 4},
		{"P2", 8},
		{"P3", 16},
		{"P4", 32},
		{"P6", 64},
		{"P10", 128},
		{"P15", 256},
		{"P20", 512},
		{"P30", 1024},
		{"P40", 2048},
		{"P50", 4096},
		{"P60", 8192},
		{"P70", 16384},
		{"P80", 32767},
	},
}

var diskProductNameMap = map[string]string{
	"Standard_LRS":    "Standard HDD Managed Disks",
	"StandardSSD_LRS": "Standard SSD Managed Disks",
	"Premium_LRS":     "Premium SSD Managed Disks",
}

// Parse from Terraform size value to Azure instance type value.
func parseVMSKUName(size string) string {
	s := strings.ReplaceAll(size, "Standard_", "")
	s = strings.ReplaceAll(s, "Basic_", "")
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

func mapDiskName(diskType string, sizeGB int) string {
	diskTypeMap, ok := diskSizeMap[diskType]
	if !ok {
		return ""
	}

	name := ""
	for _, v := range diskTypeMap {
		name = v.Name
		if v.Size >= sizeGB {
			break
		}
	}

	if sizeGB > diskTypeMap[len(diskTypeMap)-1].Size {
		return ""
	}

	return name
}

func osDiskSubResource(region string, d gjson.Result, u *schema.UsageData) *schema.Resource {
	diskType := d.Get("storage_account_type").String()

	diskSizeGB := 30
	if d.Get("disk_size_gb").Exists() {
		diskSizeGB = int(d.Get("disk_size_gb").Int())
	}

	diskName := mapDiskName(diskType, diskSizeGB)
	if diskName == "" {
		log.Warnf("Could not map disk type %s and size %d to disk name", diskType, diskSizeGB)
		return nil
	}

	productName, ok := diskProductNameMap[diskType]
	if !ok {
		log.Warnf("Could not map disk type %s to product name", diskType)
		return nil
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            fmt.Sprintf("Storage (%s)", diskName),
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(productName)},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s LRS", diskName))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Disks", diskName))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}

	if diskType == "Standard_LRS" || diskType == "StandardSSD_LRS" {
		var opsQty *decimal.Decimal

		if u != nil && u.Get("os_disk.disk_operations").Exists() {
			opsQty = decimalPtr(decimal.NewFromInt(u.Get("os_disk.disk_operations").Int()).Div(decimal.NewFromInt(10000)))
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Disk operations",
			Unit:            "10k operations",
			UnitMultiplier:  1,
			MonthlyQuantity: opsQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(productName)},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s LRS", diskName))},
					{Key: "meterName", Value: strPtr("Disk Operations")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: costComponents,
	}
}
