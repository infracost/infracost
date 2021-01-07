package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetDynamoDBTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_dynamodb_table",
		Notes: []string{
			"DAX is not yet supported.",
		},
		RFunc: NewDynamoDBTable,
	}
}

func NewDynamoDBTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	billingMode := d.Get("billing_mode").String()

	if billingMode == "PROVISIONED" {
		// Write capacity units (WCU)
		costComponents = append(costComponents, wcuCostComponent(d))
		// Read capacity units (RCU)
		costComponents = append(costComponents, rcuCostComponent(d))
	}

	// Infracost usage data

	if billingMode == "PAY_PER_REQUEST" {
		// Write request units (WRU)
		costComponents = append(costComponents, wruCostComponent(d, u))
		// Read request units (RRU)
		costComponents = append(costComponents, rruCostComponent(d, u))
	}

	// Data storage
	costComponents = append(costComponents, dataStorageCostComponent(d, u))
	// Continuous backups (PITR)
	costComponents = append(costComponents, continuousBackupCostComponent(d, u))
	// OnDemand backups
	costComponents = append(costComponents, onDemandBackupCostComponent(d, u))
	// Restoring tables
	costComponents = append(costComponents, restoreCostComponent(d, u))

	// Stream reads
	costComponents = append(costComponents, streamCostComponent(d, u))

	// Global tables (replica)
	subResources = append(subResources, globalTables(d, u)...)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func wcuCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity int64
	if d.Get("write_capacity").Exists() {
		quantity = d.Get("write_capacity").Int()
	}
	return &schema.CostComponent{
		Name:           "Write capacity unit (WCU)",
		Unit:           "WCU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(quantity)),
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
	var quantity int64
	if d.Get("read_capacity").Exists() {
		quantity = d.Get("read_capacity").Int()
	}
	return &schema.CostComponent{
		Name:           "Read capacity unit (RCU)",
		Unit:           "RCU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(quantity)),
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

func globalTables(d *schema.ResourceData, u *schema.UsageData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)
	billingMode := d.Get("billing_mode").String()
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			region := data.Get("region_name").String()
			name := fmt.Sprintf("Global table (%s)", region)
			var capacity int64
			if billingMode == "PROVISIONED" {
				capacity = d.Get("write_capacity").Int()
				resources = append(resources, newProvisionedDynamoDBGlobalTable(name, region, capacity))
			} else if billingMode == "PAY_PER_REQUEST" {
				if u != nil && u.Get("monthly_write_request_units").Exists() {
					capacity = u.Get("monthly_write_request_units").Int()
				}
				resources = append(resources, newOnDemandDynamoDBGlobalTable(name, region, capacity))
			}
		}
	}
	return resources
}

func newProvisionedDynamoDBGlobalTable(name string, region string, capacity int64) *schema.Resource {
	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			// Replicated write capacity units (rWCU)
			{
				Name:           "Replicated write capacity unit (rWCU)",
				Unit:           "rWCU-hours",
				UnitMultiplier: 1,
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

func newOnDemandDynamoDBGlobalTable(name string, region string, capacity int64) *schema.Resource {
	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			// Replicated write capacity units (rWRU)
			{
				Name:            "Replicated write request unit (rWRU)",
				Unit:            "rWRU-hours",
				UnitMultiplier:  1,
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonDynamoDB"),
					ProductFamily: strPtr("Amazon DynamoDB PayPerRequest Throughput"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "group", Value: strPtr("DDB-ReplicatedWriteUnits")},
					},
				},
			},
		},
	}
}

func wruCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_write_request_units").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_write_request_units").Int()))
	}
	return &schema.CostComponent{
		Name:            "Write request unit (WRU)",
		Unit:            "WRUs",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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

func rruCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_read_request_units").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_read_request_units").Int()))
	}
	return &schema.CostComponent{
		Name:            "Read request unit (RRU)",
		Unit:            "RRUs",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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

func dataStorageCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_gb_data_storage").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_data_storage").Int()))
	}
	return &schema.CostComponent{
		Name:            "Data storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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

func continuousBackupCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_gb_continuous_backup_storage").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_continuous_backup_storage").Int()))
	}
	return &schema.CostComponent{
		Name:            "Continuous backup storage (PITR)",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Database Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TimedPITRStorage-ByteHrs/")},
			},
		},
	}
}

func onDemandBackupCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_gb_on_demand_backup_storage").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_on_demand_backup_storage").Int()))
	}
	return &schema.CostComponent{
		Name:            "On-demand backup storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Amazon DynamoDB On-Demand Backup Storage"),
		},
	}
}

func restoreCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_gb_restore").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_restore").Int()))
	}
	return &schema.CostComponent{
		Name:            "Restore data size",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("Amazon DynamoDB Restore Data Size"),
		},
	}
}

func streamCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_streams_read_request_units").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_streams_read_request_units").Int()))
	}
	return &schema.CostComponent{
		Name:            "Streams read request unit (sRRU)",
		Unit:            "sRRUs",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDynamoDB"),
			ProductFamily: strPtr("API Request"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", ValueRegex: strPtr("/DDB-StreamsReadRequests/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			DescriptionRegex: strPtr("/beyond free tier/"),
		},
	}
}
