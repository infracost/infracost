package aws

import (
	"github.com/infracost/infracost/pkg/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func NewDynamoDBTable(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	billingMode := d.Get("billing_mode").String()

	// Write capacity units (WCU)
	if billingMode == "PROVISIONED" && d.Get("write_capacity").Exists() {
		if billingMode != "PROVISIONED" {
			log.Warnf("Skipping %s for %s. This attribute is only available for provisioned pricing.", "write_capacity", d.Address)
		} else {
			costComponents = append(costComponents, wcuCostComponent(d))
		}
	}
	// Read capacity units (RCU)
	if billingMode == "PROVISIONED" && d.Get("read_capacity").Exists() {
		if billingMode != "PROVISIONED" {
			log.Warnf("Skipping %s for %s. This attribute is only available for provisioned pricing.", "read_capacity", d.Address)
		} else {
			costComponents = append(costComponents, rcuCostComponent(d))
		}
	}

	// Global tables (replica)
	subResources = append(subResources, globalTables(d)...)

	// Infracost usage data

	// Write request units (WRU)
	if u != nil && u.Get("monthly_write_request_units").Exists() {
		if billingMode == "PROVISIONED" {
			log.Warnf("Skipping %s usage data for %s. This usage data is only available for on-demand pricing.", "monthly_write_request_units", d.Address)
		} else {
			costComponents = append(costComponents, wruCostComponent(d, u))
		}
	}
	// Read request units (RRU)
	if u != nil && u.Get("monthly_read_request_units").Exists() {
		if billingMode == "PROVISIONED" {
			log.Warnf("Skipping %s usage data for %s. This usage data is only available for on-demand pricing.", "monthly_read_request_units", d.Address)
		} else {
			costComponents = append(costComponents, rruCostComponent(d, u))
		}
	}
	// Data storage
	if u != nil && u.Get("monthly_gb_data_storage").Exists() {
		costComponents = append(costComponents, dataStorageCostComponent(d, u))
	}

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

func wruCostComponent(d *schema.ResourceData, u *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:            "Write request unit (WRU)",
		Unit:            "WRU",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(u.Get("monthly_write_request_units.0.value").Int())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Amazon DynamoDB PayPerRequest Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("DDB-WriteUnits")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func rruCostComponent(d *schema.ResourceData, u *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:            "Read request unit (RRU)",
		Unit:            "RRU",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(u.Get("monthly_read_request_units.0.value").Int())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Amazon DynamoDB PayPerRequest Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("DDB-ReadUnits")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func dataStorageCostComponent(d *schema.ResourceData, u *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:            "Data storage",
		Unit:            "GB-months",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_data_storage.0.value").Int())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Database Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TimedStorage-ByteHrs/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			DescriptionRegex: strPtr("/storage used beyond first/"),
		},
	}
}
