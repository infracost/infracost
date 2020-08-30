package aws

import (
	"fmt"
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func NewRDSClusterInstance(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)


	instanceType := rawValues["instance_class"].(string)

	switch rawValues["engine"].(string) {
	case "aurora", "aurora-mysql", nil:
		databaseEngine := "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine := "Aurora PostgreSQL"
	}

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonRDS"),
		ProductFamily: strPtr("Database Instance"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "instanceType", Value: strPtr(instanceType)},
			{Key: "databaseEngine", Value: strPtr(databaseEngine)},
		},
	}

	hoursPriceFilter := &resource.PriceFilter{
		purchaseOption: strPtr("on_demand"),
	}
	hours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour", hoursProductFilter, hoursPriceFilter)
	r.AddPriceComponent(hours)

	return r
}
