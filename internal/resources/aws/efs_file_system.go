package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type EFSFileSystem struct {
	Address                        string
	Region                         string
	HasLifecyclePolicy             bool
	AvailabilityZoneName           string
	ProvisionedThroughputInMBps    float64
	InfrequentAccessStorageGB      *float64 `infracost_usage:"infrequent_access_storage_gb"`
	StorageGB                      *float64 `infracost_usage:"storage_gb"`
	MonthlyInfrequentAccessReadGB  *float64 `infracost_usage:"monthly_infrequent_access_read_gb"`
	MonthlyInfrequentAccessWriteGB *float64 `infracost_usage:"monthly_infrequent_access_write_gb"`
}

func (r *EFSFileSystem) CoreType() string {
	return "EFSFileSystem"
}

func (r *EFSFileSystem) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "infrequent_access_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_infrequent_access_read_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_infrequent_access_write_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *EFSFileSystem) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EFSFileSystem) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	var storageGB *decimal.Decimal
	if r.StorageGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	if r.AvailabilityZoneName != "" {
		costComponents = append(costComponents, r.storageCostComponent("Storage (one zone)", "-TimedStorage-Z-ByteHrs", storageGB))
	} else {
		costComponents = append(costComponents, r.storageCostComponent("Storage (standard)", "-TimedStorage-ByteHrs", storageGB))
	}

	if r.ProvisionedThroughputInMBps > 0 {
		provisionedThroughput := r.calculateProvisionedThroughput(storageGB, decimal.NewFromFloat(r.ProvisionedThroughputInMBps))
		costComponents = append(costComponents, r.provisionedThroughputCostComponent(provisionedThroughput))
	}

	if r.HasLifecyclePolicy {
		var infrequentAccessStorageGB *decimal.Decimal
		if r.InfrequentAccessStorageGB != nil {
			infrequentAccessStorageGB = decimalPtr(decimal.NewFromFloat(*r.InfrequentAccessStorageGB))
		}

		if r.AvailabilityZoneName != "" {
			costComponents = append(costComponents, r.storageCostComponent("Storage (one zone, infrequent access)", "IATimedStorage-Z-ByteHrs", infrequentAccessStorageGB))
		} else {
			costComponents = append(costComponents, r.storageCostComponent("Storage (standard, infrequent access)", "-IATimedStorage-ByteHrs", infrequentAccessStorageGB))
		}

		var infrequentAccessReadGB *decimal.Decimal
		if r.MonthlyInfrequentAccessReadGB != nil {
			infrequentAccessReadGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyInfrequentAccessReadGB))
		}

		var infrequentAccessWriteGB *decimal.Decimal
		if r.MonthlyInfrequentAccessWriteGB != nil {
			infrequentAccessWriteGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyInfrequentAccessWriteGB))
		}

		costComponents = append(costComponents, r.requestsCostComponent("Read requests (infrequent access)", "Read", infrequentAccessReadGB))
		costComponents = append(costComponents, r.requestsCostComponent("Write requests (infrequent access)", "Write", infrequentAccessWriteGB))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *EFSFileSystem) calculateProvisionedThroughput(storageGB *decimal.Decimal, throughput decimal.Decimal) *decimal.Decimal {
	if storageGB == nil {
		storageGB = &decimal.Zero
	}

	defaultThroughput := storageGB.Mul(decimal.NewFromInt(730).Div(decimal.NewFromInt(20).Mul(decimal.NewFromInt(1))))
	totalProvisionedThroughput := throughput.Mul(decimal.NewFromInt(730))
	totalBillableProvisionedThroughput := totalProvisionedThroughput.Sub(defaultThroughput).Div(decimal.NewFromInt(730))

	if totalBillableProvisionedThroughput.IsPositive() {
		return &totalBillableProvisionedThroughput
	}

	return &decimal.Zero
}

func (r *EFSFileSystem) storageCostComponent(name, usagetype string, storageGB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonEFS"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usagetype))},
			},
		},
		UsageBased: true,
	}
}

func (r *EFSFileSystem) provisionedThroughputCostComponent(provisionedThroughputMiBps *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Provisioned throughput",
		Unit:            "MBps",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: provisionedThroughputMiBps,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonEFS"),
			ProductFamily: strPtr("Provisioned Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ProvisionedTP-MiBpsHrs/")},
			},
		},
		UsageBased: true,
	}
}

func (r *EFSFileSystem) requestsCostComponent(name, accessType string, requestsGB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: requestsGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonEFS"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "accessType", Value: strPtr(accessType)},
				{Key: "storageClass", Value: strPtr("Infrequent Access")},
			},
		},
		UsageBased: true,
	}
}
