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

func (r *DMSReplicationInstance) CoreType() string {
	return "DMSReplicationInstance"
}

func (r *DMSReplicationInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *DMSReplicationInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DMSReplicationInstance) BuildResource() *schema.Resource {
	instanceTypeParts := strings.Split(r.ReplicationInstanceClass, ".")
	if len(instanceTypeParts) < 3 {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}
	instanceType := strings.Join(instanceTypeParts[1:], ".")
	instanceFamily := instanceTypeParts[1]

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, r.instanceCostComponent(instanceType))
	costComponents = append(costComponents, r.storageCostComponent(instanceFamily))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *DMSReplicationInstance) instanceCostComponent(instanceType string) *schema.CostComponent {
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

func (r *DMSReplicationInstance) storageCostComponent(instanceFamily string) *schema.CostComponent {
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
