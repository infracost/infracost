package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDMSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dms_replication_instance",
		RFunc: NewDMSReplicationInstance,
	}
}

func getInstanceType(d *schema.ResourceData) string {
	rawInstanceType := strings.Split(d.Get("replication_instance_class").String(), ".")
	instanceType := strings.Join(rawInstanceType[1:], ".")
	return instanceType
}

func getInstanceFamily(d *schema.ResourceData) string {
	rawInstanceType := strings.Split(d.Get("replication_instance_class").String(), ".")
	return rawInstanceType[1]
}

func instanceCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	instanceType := getInstanceType(d)
	availabilityZone := "Single"
	if d.Get("multi_az").Bool() {
		availabilityZone = "Multiple"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance (%s)", instanceType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AWSDatabaseMigrationSvc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(instanceType)},
				{Key: "availabilityZone", Value: strPtr(availabilityZone)},
			},
		},
	}
}

func storageCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	instanceFamily := getInstanceFamily(d)
	availabilityZone := "Single"
	if d.Get("multi_az").Bool() {
		availabilityZone = "Multiple"
	}

	baseStorageSize := d.Get("allocated_storage").Int()
	var freeStorageSize int64 = 0
	switch instanceFamily {
	case "c4":
		freeStorageSize = 100
	case "r4":
		freeStorageSize = 100
	case "r5":
		freeStorageSize = 100
	case "t2":
		freeStorageSize = 50
	case "t3":
		freeStorageSize = 50
	}
	var storageSize int64 = 0
	if baseStorageSize > freeStorageSize {
		storageSize = baseStorageSize - freeStorageSize
	}

	return &schema.CostComponent{
		Name:            "Storage (general purpose SSD, gp2)",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(storageSize)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AWSDatabaseMigrationSvc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "storageMedia", Value: strPtr("SSD")},
				{Key: "availabilityZone", Value: strPtr(availabilityZone)},
			},
		},
	}
}

func NewDMSReplicationInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, instanceCostComponent(d))
	costComponents = append(costComponents, storageCostComponent(d))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
