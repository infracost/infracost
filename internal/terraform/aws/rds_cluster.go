package aws

import (
	"infracost/pkg/resource"
)

func NewRDSCluster(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	var databaseEngine string
	switch rawValues["engine"].(string) {
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
	}

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonRDS"),
		ProductFamily: strPtr("Aurora Global Database"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "databaseEngine", ValueRegex: strPtr(databaseEngine)},
		},
	}
	hours := resource.NewBasePriceComponent("hours", r, "hour", "hour", hoursProductFilter, nil)
	r.AddPriceComponent(hours)

	return r
}
