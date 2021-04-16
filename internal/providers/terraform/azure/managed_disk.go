package azure

import (
	"fmt"
	"math"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var diskSizeMap = map[string][]struct {
	Name string
	Size int
}{
	// The mapping is from https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types
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

var ultraDiskSizes = []int{4, 8, 16, 32, 64, 128, 256, 512}
var ultraDiskSizeStep = 1024
var ultraDiskMaxSize = 65536

var diskProductNameMap = map[string]string{
	"Standard_LRS":    "Standard HDD Managed Disks",
	"StandardSSD_LRS": "Standard SSD Managed Disks",
	"Premium_LRS":     "Premium SSD Managed Disks",
}

func GetAzureRMManagedDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_managed_disk",
		RFunc: NewAzureRMManagedDisk,
	}
}

func NewAzureRMManagedDisk(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()

	var monthlyDiskOperations *decimal.Decimal

	if u != nil && u.Get("monthly_disk_operations").Exists() {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_disk_operations").Int()))
	}

	costComponents := managedDiskCostComponents(region, d.RawValues, monthlyDiskOperations)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func managedDiskCostComponents(region string, diskData gjson.Result, monthlyDiskOperations *decimal.Decimal) []*schema.CostComponent {
	diskType := diskData.Get("storage_account_type").String()

	if diskType == "UltraSSD_LRS" {
		return ultraDiskCostComponents(region, diskType, diskData)
	}

	return standardPremiumDiskCostComponents(region, diskType, diskData, monthlyDiskOperations)
}

func standardPremiumDiskCostComponents(region string, diskType string, diskData gjson.Result, monthlyDiskOperations *decimal.Decimal) []*schema.CostComponent {
	requestedSize := 30

	if diskData.Get("disk_size_gb").Exists() {
		requestedSize = int(diskData.Get("disk_size_gb").Int())
	}

	diskName := mapDiskName(diskType, requestedSize)
	if diskName == "" {
		log.Warnf("Could not map disk type %s and size %d to disk name", diskType, requestedSize)
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

		if monthlyDiskOperations != nil {
			opsQty = decimalPtr(monthlyDiskOperations.Div(decimal.NewFromInt(10000)))
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

	return costComponents
}

func ultraDiskCostComponents(region string, diskType string, diskData gjson.Result) []*schema.CostComponent {
	requestedSize := 1024
	iops := 2048
	throughput := 8

	if diskData.Get("disk_size_gb").Exists() {
		requestedSize = int(diskData.Get("disk_size_gb").Int())
	}

	if diskData.Get("disk_iops_read_write").Exists() {
		iops = int(diskData.Get("disk_iops_read_write").Int())
	}

	if diskData.Get("disk_mbps_read_write").Exists() {
		throughput = int(diskData.Get("disk_mbps_read_write").Int())
	}

	diskSize := mapUltraDiskSize(requestedSize)

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Storage (ultra, %d GiB)", diskSize),
			Unit:           "GiB-hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(diskSize))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr("Ultra LRS")},
					{Key: "meterName", Value: strPtr("Provisioned Capacity")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
		{
			Name:           "Provisioned IOPS",
			Unit:           "IOPS-hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(iops))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr("Ultra LRS")},
					{Key: "meterName", Value: strPtr("Provisioned IOPS")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
		{
			Name:           "Throughput",
			Unit:           "MB/s-hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(throughput))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr("Ultra LRS")},
					{Key: "meterName", Value: strPtr("Provisioned Throughput (MBps)")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}

	return costComponents
}

func mapDiskName(diskType string, requestedSize int) string {
	diskTypeMap, ok := diskSizeMap[diskType]
	if !ok {
		return ""
	}

	name := ""
	for _, v := range diskTypeMap {
		name = v.Name
		if v.Size >= requestedSize {
			break
		}
	}

	if requestedSize > diskTypeMap[len(diskTypeMap)-1].Size {
		return ""
	}

	return name
}

func mapUltraDiskSize(requestedSize int) int {
	if requestedSize >= ultraDiskMaxSize {
		return ultraDiskMaxSize
	}

	if requestedSize < ultraDiskSizes[0] {
		return ultraDiskSizes[0]
	}

	if requestedSize > ultraDiskSizes[len(ultraDiskSizes)-1] {
		return int(math.Ceil(float64(requestedSize)/float64(ultraDiskSizeStep))) * ultraDiskSizeStep
	}

	size := 0
	for _, v := range ultraDiskSizes {
		size = v
		if size >= requestedSize {
			break
		}
	}

	return size

}
