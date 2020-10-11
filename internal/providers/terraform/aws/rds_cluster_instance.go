package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster_instance",
		RFunc: NewRDSClusterInstance,
	}
}

func NewRDSClusterInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	instanceType := d.Get("instance_class").String()

	var databaseEngine *string
	switch d.Get("engine").String() {
	case "aurora", "aurora-mysql", "":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Database instance",
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonRDS"),
					ProductFamily: strPtr("Database Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "instanceType", Value: strPtr(instanceType)},
						{Key: "databaseEngine", Value: databaseEngine},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("on_demand"),
				},
			},
		},
	}
}
