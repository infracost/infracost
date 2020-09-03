package aws

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func NewDynamoDBTable(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	// Check billing mode
	billingMode := d.Get("billing_mode").String()
	if billingMode != "PROVISIONED" {
		log.Warnf("Skipping resource %s. Infracost currently only supports the PROVISIONED billing mode for AWS DynamoDB tables", d.Address)
		return nil
	}

	// Write capacity units (WCU)
	costComponents = append(costComponents, wcuCostComponent(d))
	// Read capacity units (RCU)
	costComponents = append(costComponents, rcuCostComponent(d))

	// Global tables (replica)
	subResources = append(subResources, globalTables(d)...)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func wcuCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:           "Write capacity unit (WCU)",
		Unit:           "WCU/hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(d.Get("write_capacity").Int())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Provisioned IOPS"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("DDB-WriteUnits")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			DescriptionRegex: strPtr("/beyond the free tier/"),
		},
	}
}

func rcuCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:           "Read capacity unit (RCU)",
		Unit:           "RCU/hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(d.Get("read_capacity").Int())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Provisioned IOPS"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("DDB-ReadUnits")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			DescriptionRegex: strPtr("/beyond the free tier/"),
		},
	}
}

func globalTables(d *schema.ResourceData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			region := data.Get("region_name").String()
			name := region
			capacity := d.Get("write_capacity").Int()
			resources = append(resources, newDynamoDBGlobalTable(name, data, region, capacity))
		}
	}
	return resources
}

func newDynamoDBGlobalTable(name string, d gjson.Result, region string, capacity int64) *schema.Resource {
	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			// Replicated write capacity units (rWCU)
			{
				Name:           "Replicated write capacity unit (rWCU)",
				Unit:           "rWCU/hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonDynamoDB"),
					ProductFamily: strPtr("DDB-Operation-ReplicatedWrite"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "group", Value: strPtr("DDB-ReplicatedWriteUnits")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption:   strPtr("on_demand"),
					DescriptionRegex: strPtr("/beyond the free tier/"),
				},
			},
		},
	}
}
