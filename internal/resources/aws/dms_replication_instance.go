package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DMSReplicationInstance struct {
	Address                  string
	Region                   string
	AllocatedStorageGB       int64
	ReplicationInstanceClass string
	MultiAZ                  bool
}

var DMSReplicationInstanceUsageSchema = []*schema.UsageItem{}

func (r *DMSReplicationInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DMSReplicationInstance) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, r.instanceCostComponent())
	costComponents = append(costComponents, r.storageCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    DMSReplicationInstanceUsageSchema,
	}
}
func (r *DMSReplicationInstance) getInstanceType() string {
	rawInstanceType := strings.Split(r.ReplicationInstanceClass, ".")
	instanceType := strings.Join(rawInstanceType[1:], ".")
	return instanceType
}

func (r *DMSReplicationInstance) getInstanceFamily() string {
	rawInstanceType := strings.Split(r.ReplicationInstanceClass, ".")
	return rawInstanceType[1]
}

func (r *DMSReplicationInstance) instanceCostComponent() *schema.CostComponent {
	instanceType := r.getInstanceType()
	availabilityZone := "Single"
	if r.MultiAZ {
		availabilityZone = "Multiple"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance (%s)", instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AWSDatabaseMigrationSvc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(instanceType)},
				{Key: "availabilityZone", Value: strPtr(availabilityZone)},
			},
		},
	}
}

func (r *DMSReplicationInstance) storageCostComponent() *schema.CostComponent {
	instanceFamily := r.getInstanceFamily()
	availabilityZone := "Single"
	if r.MultiAZ {
		availabilityZone = "Multiple"
	}

	baseStorageSize := r.AllocatedStorageGB
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
			Region:     strPtr(r.Region),
			Service:    strPtr("AWSDatabaseMigrationSvc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "storageMedia", Value: strPtr("SSD")},
				{Key: "availabilityZone", Value: strPtr(availabilityZone)},
			},
		},
	}
}
