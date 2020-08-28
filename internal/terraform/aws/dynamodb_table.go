package aws

import (
	"fmt"
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func NewDynamoDBGlobalTable(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonDynamoDB"),
		ProductFamily: strPtr("DDB-Operation-ReplicatedWrite"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "group", Value: strPtr("DDB-ReplicatedWriteUnits")},
		},
	}
	hoursPriceFilter := &resource.PriceFilter{
		PurchaseOption:   strPtr("on_demand"),
		DescriptionRegex: strPtr("/beyond the free tier/"),
	}
	rwcu := resource.NewBasePriceComponent("Replicated write capacity unit (rWCU)", r, "rWCU/hour", "hour", hoursProductFilter, hoursPriceFilter)
	rwcu.SetQuantityMultiplierFunc(dynamoDBWCUQuantity)
	r.AddPriceComponent(rwcu)

	return r
}

func dynamoDBWCUQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	capacity := resource.RawValues()["write_capacity"]
	if capacity != nil {
		quantity = decimal.NewFromFloat((capacity.(float64)))
	}
	return quantity
}

func dynamoDBRCUQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	capacity := resource.RawValues()["read_capacity"]
	if capacity != nil {
		quantity = decimal.NewFromFloat((capacity.(float64)))
	}
	return quantity
}

func NewDynamoDBTable(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	// Check billing mode
	billingMode := rawValues["billing_mode"]
	if billingMode != "PROVISIONED" {
		log.Warnf("No support for on-demand dynamoDB for %s", address)
		return r
	}

	// Write capacity unit (WCU)
	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonDynamoDB"),
		ProductFamily: strPtr("Provisioned IOPS"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "group", Value: strPtr("DDB-WriteUnits")},
		},
	}
	hoursPriceFilter := &resource.PriceFilter{
		PurchaseOption:   strPtr("on_demand"),
		DescriptionRegex: strPtr("/beyond the free tier/"),
	}
	wcu := resource.NewBasePriceComponent("Write capacity unit (WCU)", r, "WCU/hour", "hour", hoursProductFilter, hoursPriceFilter)
	wcu.SetQuantityMultiplierFunc(dynamoDBWCUQuantity)
	r.AddPriceComponent(wcu)

	// Read capacity unit (RCU)
	hoursProductFilter = &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonDynamoDB"),
		ProductFamily: strPtr("Provisioned IOPS"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "group", Value: strPtr("DDB-ReadUnits")},
		},
	}
	hoursPriceFilter = &resource.PriceFilter{
		PurchaseOption:   strPtr("on_demand"),
		DescriptionRegex: strPtr("/beyond the free tier/"),
	}
	rcu := resource.NewBasePriceComponent("Read capacity unit (RCU)", r, "RCU/hour", "hour", hoursProductFilter, hoursPriceFilter)
	rcu.SetQuantityMultiplierFunc(dynamoDBRCUQuantity)
	r.AddPriceComponent(rcu)

	// Global table (replica)
	if r.RawValues()["replica"] != nil {
		replicasRawValues := r.RawValues()["replica"].([]interface{})
		for _, replicaRawValues := range replicasRawValues {
			rawValues := replicaRawValues.(map[string]interface{})
			rawValues["write_capacity"] = r.RawValues()["write_capacity"]
			replicaRegion := rawValues["region_name"].(string)
			replicaAddress := fmt.Sprintf("%s.global_table.%s", r.Address(), replicaRegion)
			replica := NewDynamoDBGlobalTable(replicaAddress, replicaRegion, rawValues)
			r.AddSubResource(replica)
		}
	}

	return r
}
