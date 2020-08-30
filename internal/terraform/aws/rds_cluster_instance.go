package aws

import (
	"infracost/pkg/resource"
)

func NewRDSClusterInstance(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	instanceType := rawValues["instance_class"].(string)

	var databaseEngine string
	switch rawValues["engine"].(string) {
	case "aurora", "aurora-mysql", "":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
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
		PurchaseOption: strPtr("on_demand"),
	}
	hours := resource.NewBasePriceComponent("instance hours", r, "hour", "hour", hoursProductFilter, hoursPriceFilter)
	r.AddPriceComponent(hours)

	return r
}
