package azure

import (
	"slices"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"math"
	"strings"

	"github.com/shopspring/decimal"
)

const Standard = "Standard"
const StandardSSD = "StandardSSD"
const Premium = "Premium"

type ManagedDisk struct {
	Address string
	Region  string
	ManagedDiskData
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

type ManagedDiskData struct {
	DiskType          string
	DiskSizeGB        int64
	DiskIOPSReadWrite int64
	DiskMBPSReadWrite int64
}

func (r *ManagedDisk) CoreType() string {
	return "ManagedDisk"
}

func (r *ManagedDisk) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_disk_operations", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *ManagedDisk) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ManagedDisk) BuildResource() *schema.Resource {
	region := r.Region
	diskType := r.DiskType

	var monthlyDiskOperations *decimal.Decimal

	if r.MonthlyDiskOperations != nil {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(*r.MonthlyDiskOperations))
	}

	costComponents := managedDiskCostComponents(region, diskType, r.DiskSizeGB, r.DiskIOPSReadWrite, r.DiskMBPSReadWrite, monthlyDiskOperations)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

var diskSizeMap = map[string][]struct {
	Name string
	Size int
}{

	"Standard": {
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
	"StandardSSD": {
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
	"Premium": {
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

var storageReplicationTypes = []string{"LRS", "ZRS"}
var ultraDiskSizes = []int{4, 8, 16, 32, 64, 128, 256, 512}
var ultraDiskSizeStep = 1024
var ultraDiskMaxSize = 65536

var diskProductNameMap = map[string]string{
	"Standard":    "Standard HDD Managed Disks",
	"StandardSSD": "Standard SSD Managed Disks",
	"Premium":     "Premium SSD Managed Disks",
}

func managedDiskCostComponents(region, diskType string, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite int64, monthlyDiskOperations *decimal.Decimal) []*schema.CostComponent {
	p := strings.Split(diskType, "_")
	diskTypePrefix := p[0]

	var storageReplicationType string
	if len(p) > 1 {
		storageReplicationType = strings.ToUpper(p[1])
	}

	validstorageReplicationType := mapStorageReplicationType(storageReplicationType)
	if !validstorageReplicationType {
		logging.Logger.Warn().Msgf("Could not map %s to a valid storage type", storageReplicationType)
		return nil
	}

	if strings.ToLower(diskTypePrefix) == "ultrassd" {
		return ultraDiskCostComponents(region, storageReplicationType, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite)
	}

	return standardPremiumDiskCostComponents(region, diskTypePrefix, storageReplicationType, diskSizeGB, monthlyDiskOperations)
}

func standardPremiumDiskCostComponents(region string, diskTypePrefix string, storageReplicationType string, diskSizeGB int64, monthlyDiskOperations *decimal.Decimal) []*schema.CostComponent {
	requestedSize := 30

	if diskSizeGB > 0 {
		requestedSize = int(diskSizeGB)
	}

	diskName := mapDiskName(diskTypePrefix, requestedSize)
	if diskName == "" {
		logging.Logger.Warn().Msgf("Could not map disk type %s and size %d to disk name", diskTypePrefix, requestedSize)
		return nil
	}

	productName, ok := diskProductNameMap[diskTypePrefix]
	if !ok {
		logging.Logger.Warn().Msgf("Could not map disk type %s to product name", diskTypePrefix)
		return nil
	}

	costComponents := []*schema.CostComponent{storageCostComponent(region, diskName, storageReplicationType, productName)}

	if strings.ToLower(diskTypePrefix) == "standard" || strings.ToLower(diskTypePrefix) == "standardssd" {
		var opsQty *decimal.Decimal

		if monthlyDiskOperations != nil {
			opsQty = decimalPtr(monthlyDiskOperations.Div(decimal.NewFromInt(10000)))
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Disk operations",
			Unit:            "10k operations",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: opsQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(productName)},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s %s", diskName, storageReplicationType))},
					{Key: "meterName", ValueRegex: regexPtr("Disk Operations$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
			UsageBased: true,
		})
	}

	return costComponents
}

func storageCostComponent(region, diskName, storageReplicationType, productName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", diskName, storageReplicationType),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%s %s", diskName, storageReplicationType))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("^%s (%s )?Disk(s)?$", diskName, storageReplicationType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func ultraDiskCostComponents(region string, storageReplicationType string, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite int64) []*schema.CostComponent {
	requestedSize := 1024
	iops := 2048
	throughput := 8

	if diskSizeGB > 0 {
		requestedSize = int(diskSizeGB)
	}

	if diskIOPSReadWrite > 0 {
		iops = int(diskIOPSReadWrite)
	}

	if diskMBPSReadWrite > 0 {
		throughput = int(diskMBPSReadWrite)
	}

	diskSize := mapUltraDiskSize(requestedSize)

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Storage (ultra, %d GiB)", diskSize),
			Unit:           "GiB",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(diskSize))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("Ultra %s", storageReplicationType))},
					{Key: "meterName", ValueRegex: regexPtr("Provisioned Capacity$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
		{
			Name:           "Provisioned IOPS",
			Unit:           "IOPS",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(iops))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("Ultra %s", storageReplicationType))},
					{Key: "meterName", ValueRegex: regexPtr("Provisioned IOPS$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
		{
			Name:           "Throughput",
			Unit:           "MB/s",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(throughput))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Ultra Disks")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("Ultra %s", storageReplicationType))},
					{Key: "meterName", ValueRegex: regexPtr("Provisioned Throughput \\(MBps\\)$")},
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

func mapStorageReplicationType(storageReplicationType string) bool {
	return slices.Contains(storageReplicationTypes, storageReplicationType)
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
