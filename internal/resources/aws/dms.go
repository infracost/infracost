package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DMS struct {
	Address                  *string
	Region                   *string
	MultiAz                  *bool
	AllocatedStorage         *int64
	ReplicationInstanceClass *string
}

var DMSUsageSchema = []*schema.UsageItem{}

func (r *DMS) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DMS) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, instanceCostComponent(r))
	costComponents = append(costComponents, storageCostComponent(r))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: DMSUsageSchema,
	}
}
func getInstanceType(r *DMS,) string {
	rawInstanceType := strings.Split(*r.ReplicationInstanceClass, ".")
	instanceType := strings.Join(rawInstanceType[1:], ".")
	return instanceType
}

func getInstanceFamily(r *DMS,) string {
	rawInstanceType := strings.Split(*r.ReplicationInstanceClass, ".")
	return rawInstanceType[1]
}

func instanceCostComponent(r *DMS,) *schema.CostComponent {
	region := *r.Region
	instanceType := getInstanceType(r)
	availabilityZone := "Single"
	if *r.MultiAz {
		availabilityZone = "Multiple"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance (%s)", instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
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

func storageCostComponent(r *DMS,) *schema.CostComponent {
	region := *r.Region
	instanceFamily := getInstanceFamily(r)
	availabilityZone := "Single"
	if *r.MultiAz {
		availabilityZone = "Multiple"
	}

	baseStorageSize := *r.AllocatedStorage
	var freeStorageSize int64
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
	var storageSize int64
	if baseStorageSize > freeStorageSize {
		storageSize = baseStorageSize - freeStorageSize
	}

	return &schema.CostComponent{
		Name:            "Storage (general purpose SSD, gp2)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
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
