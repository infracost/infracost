package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type DynamoDbTableArguments struct {
	Address        string   `json:"address,omitempty"`
	Region         string   `json:"region,omitempty"`
	BillingMode    string   `json:"billingMode,omitempty"`
	WriteCapacity  int64    `json:"writeCapacity,omitempty"`
	ReadCapacity   int64    `json:"readCapacity,omitempty"`
	ReplicaRegions []string `json:"replicaRegions,omitempty"`

	MonthlyWriteRequestUnits       *int64 `json:"monthlyWriteRequestUnits,omitempty"`
	MonthlyReadRequestUnits        *int64 `json:"monthlyReadRequestUnits,omitempty"`
	StorageGB                      *int64 `json:"storageGB,omitempty"`
	PitrBackupStorageGB            *int64 `json:"pitrBackupStorageGB,omitempty"`
	OnDemandBackupStorageGB        *int64 `json:"onDemandBackupStorageGB,omitempty"`
	MonthlyDataRestoredGB          *int64 `json:"monthlyDataRestoredGB,omitempty"`
	MonthlyStreamsReadRequestUnits *int64 `json:"monthlyStreamsReadRequestUnits,omitempty"`
}

func (args *DynamoDbTableArguments) PopulateUsage(u *schema.UsageData) {
	if u != nil {
		args.MonthlyWriteRequestUnits = u.GetInt("monthly_write_request_units")
		args.MonthlyReadRequestUnits = u.GetInt("monthly_read_request_units")
		args.StorageGB = u.GetInt("storage_gb")
		args.PitrBackupStorageGB = u.GetInt("pitr_backup_storage_gb")
		args.OnDemandBackupStorageGB = u.GetInt("on_demand_backup_storage_gb")
		args.MonthlyDataRestoredGB = u.GetInt("monthly_data_restored_gb")
		args.MonthlyStreamsReadRequestUnits = u.GetInt("monthly_streams_read_request_units")
	}
}

var DynamoDbTableUsageSchema = []*schema.UsageSchemaItem{
	{Key: "monthly_write_request_units", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "pitr_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "on_demand_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_data_restored_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_streams_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
}

func NewDynamoDBTable(args *DynamoDbTableArguments) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	if args.BillingMode == "PROVISIONED" {
		// Write capacity units (WCU)
		costComponents = append(costComponents, wcuCostComponent(args.Region, args.WriteCapacity))
		// Read capacity units (RCU)
		costComponents = append(costComponents, rcuCostComponent(args.Region, args.ReadCapacity))
	}

	// Infracost usage data

	if args.BillingMode == "PAY_PER_REQUEST" {
		// Write request units (WRU)
		costComponents = append(costComponents, wruCostComponent(args.Region, args.MonthlyWriteRequestUnits))
		// Read request units (RRU)
		costComponents = append(costComponents, rruCostComponent(args.Region, args.MonthlyReadRequestUnits))
	}

	// Data storage
	costComponents = append(costComponents, dataStorageCostComponent(args.Region, args.StorageGB))
	// Continuous backups (PITR)
	costComponents = append(costComponents, continuousBackupCostComponent(args.Region, args.PitrBackupStorageGB))
	// OnDemand backups
	costComponents = append(costComponents, onDemandBackupCostComponent(args.Region, args.OnDemandBackupStorageGB))
	// Restoring tables
	costComponents = append(costComponents, restoreCostComponent(args.Region, args.MonthlyDataRestoredGB))

	// Stream reads
	costComponents = append(costComponents, streamCostComponent(args.Region, args.MonthlyStreamsReadRequestUnits))

	// Global tables (replica)
	subResources = append(subResources, globalTables(args.BillingMode, args.ReplicaRegions, args.WriteCapacity, args.MonthlyWriteRequestUnits)...)

	return &schema.Resource{
		Name:           args.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func wcuCostComponent(region string, writeCapacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Write capacity unit (WCU)",
		Unit:           "WCU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(writeCapacityUnits)),
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

func rcuCostComponent(region string, readCapacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Read capacity unit (RCU)",
		Unit:           "RCU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(readCapacityUnits)),
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

func globalTables(billingMode string, replicaRegions []string, writeCapacity int64, monthlyWRU *int64) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for _, region := range replicaRegions {
		name := fmt.Sprintf("Global table (%s)", region)
		var capacity int64
		if billingMode == "PROVISIONED" {
			capacity = writeCapacity
			resources = append(resources, newProvisionedDynamoDBGlobalTable(name, region, capacity))
		} else if billingMode == "PAY_PER_REQUEST" {
			if monthlyWRU != nil {
				capacity = *monthlyWRU
			}
			resources = append(resources, newOnDemandDynamoDBGlobalTable(name, region, capacity))
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
				Unit:           "rWCU",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
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
				Unit:            "rWRU",
				UnitMultiplier:  schema.HourToMonthUnitMultiplier,
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

func wruCostComponent(region string, monthlyWRU *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if monthlyWRU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*monthlyWRU))
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

func rruCostComponent(region string, monthlyRRU *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if monthlyRRU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*monthlyRRU))
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

func dataStorageCostComponent(region string, storageGB *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if storageGB != nil {
		quantity = decimalPtr(decimal.NewFromInt(*storageGB))
	}
	return &schema.CostComponent{
		Name:            "Data storage",
		Unit:            "GB",
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

func continuousBackupCostComponent(region string, pitrBackupStorageGB *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if pitrBackupStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromInt(*pitrBackupStorageGB))
	}
	return &schema.CostComponent{
		Name:            "Point-In-Time Recovery (PITR) backup storage",
		Unit:            "GB",
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

func onDemandBackupCostComponent(region string, onDemandBackupStorageGB *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if onDemandBackupStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromInt(*onDemandBackupStorageGB))
	}
	return &schema.CostComponent{
		Name:            "On-demand backup storage",
		Unit:            "GB",
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

func restoreCostComponent(region string, monthlyDataRestoredGB *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if monthlyDataRestoredGB != nil {
		quantity = decimalPtr(decimal.NewFromInt(*monthlyDataRestoredGB))
	}
	return &schema.CostComponent{
		Name:            "Table data restored",
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

func streamCostComponent(region string, monthlyStreamsRRU *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if monthlyStreamsRRU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*monthlyStreamsRRU))
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
