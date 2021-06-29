package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type DynamoDbTableArguments struct {
	Address string `json:"address,omitempty"`

	// "required" args that can't really be missing.  These can still be overridden by usage.
	Region         string   `json:"region,omitempty"`
	BillingMode    string   `json:"billingMode,omitempty"`
	ReplicaRegions []string `json:"replicaRegions,omitempty"`

	// "optional" args, meaning that we may show a 'Monthly cost depends on usage...' if they are missing.
	WriteCapacity                  *int64 `json:"writeCapacity,omitempty"`
	ReadCapacity                   *int64 `json:"readCapacity,omitempty"`
	MonthlyWriteRequestUnits       *int64 `json:"monthlyWriteRequestUnits,omitempty"`
	MonthlyReadRequestUnits        *int64 `json:"monthlyReadRequestUnits,omitempty"`
	StorageGB                      *int64 `json:"storageGB,omitempty"`
	PitrBackupStorageGB            *int64 `json:"pitrBackupStorageGB,omitempty"`
	OnDemandBackupStorageGB        *int64 `json:"onDemandBackupStorageGB,omitempty"`
	MonthlyDataRestoredGB          *int64 `json:"monthlyDataRestoredGB,omitempty"`
	MonthlyStreamsReadRequestUnits *int64 `json:"monthlyStreamsReadRequestUnits,omitempty"`
}

func (args *DynamoDbTableArguments) buildUsageSchema(keysToSkipSync []string) []*schema.UsageSchemaItem {
	schema := []*schema.UsageSchemaItem{
		{Key: "region", DefaultValue: "us-east-1", ValueType: schema.String},
		{Key: "billing_mode", DefaultValue: "PROVISIONED", ValueType: schema.String},
		{Key: "replica_regions", DefaultValue: []string{}, ValueType: schema.StringArray},
		{Key: "write_capacity", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "read_capacity", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_write_request_units", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "pitr_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "on_demand_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_restored_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_streams_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
	}
	setShouldSync(schema, keysToSkipSync)
	return schema
}

func (args *DynamoDbTableArguments) populateFromUsage(u *schema.UsageData) {
	if u != nil {
		if val := u.GetString("region"); val != nil {
			args.Region = *val
		}
		if val := u.GetString("billing_mode"); val != nil {
			args.BillingMode = *val
		}
		if val := u.GetStringArray("replica_regions"); val != nil {
			args.ReplicaRegions = *val
		}
		if val := u.GetInt("write_capacity"); val != nil {
			args.WriteCapacity = val
		}
		if val := u.GetInt("read_capacity"); val != nil {
			args.ReadCapacity = val
		}
		if val := u.GetInt("monthly_write_request_units"); val != nil {
			args.MonthlyWriteRequestUnits = val
		}
		if val := u.GetInt("monthly_read_request_units"); val != nil {
			args.MonthlyReadRequestUnits = val
		}
		if val := u.GetInt("storage_gb"); val != nil {
			args.StorageGB = val
		}
		if val := u.GetInt("pitr_backup_storage_gb"); val != nil {
			args.PitrBackupStorageGB = val
		}
		if val := u.GetInt("on_demand_backup_storage_gb"); val != nil {
			args.OnDemandBackupStorageGB = val
		}
		if val := u.GetInt("monthly_data_restored_gb"); val != nil {
			args.MonthlyDataRestoredGB = val
		}
		if val := u.GetInt("monthly_streams_read_request_units"); val != nil {
			args.MonthlyStreamsReadRequestUnits = val
		}
	}
}

func NewDynamoDBTable(args *DynamoDbTableArguments, u *schema.UsageData, keysToSkipSync []string) *schema.Resource {
	args.populateFromUsage(u)

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
		UsageSchema:    args.buildUsageSchema(keysToSkipSync),
	}
}

func wcuCostComponent(region string, provisionedWCU *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if provisionedWCU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*provisionedWCU))
	}
	return &schema.CostComponent{
		Name:           "Write capacity unit (WCU)",
		Unit:           "WCU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
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

func rcuCostComponent(region string, provisionedRCU *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if provisionedRCU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*provisionedRCU))
	}
	return &schema.CostComponent{
		Name:           "Read capacity unit (RCU)",
		Unit:           "RCU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
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

func globalTables(billingMode string, replicaRegions []string, writeCapacity *int64, monthlyWRU *int64) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for _, region := range replicaRegions {
		name := fmt.Sprintf("Global table (%s)", region)
		if billingMode == "PROVISIONED" {
			resources = append(resources, newProvisionedDynamoDBGlobalTable(name, region, writeCapacity))
		} else if billingMode == "PAY_PER_REQUEST" {
			resources = append(resources, newOnDemandDynamoDBGlobalTable(name, region, monthlyWRU))
		}
	}

	return resources
}

func newProvisionedDynamoDBGlobalTable(name string, region string, provisionedWCU *int64) *schema.Resource {
	var quantity *decimal.Decimal
	if provisionedWCU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*provisionedWCU))
	}

	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			// Replicated write capacity units (rWCU)
			{
				Name:           "Replicated write capacity unit (rWCU)",
				Unit:           "rWCU",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
				HourlyQuantity: quantity,
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

func newOnDemandDynamoDBGlobalTable(name string, region string, monthlyWRU *int64) *schema.Resource {
	var quantity *decimal.Decimal
	if monthlyWRU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*monthlyWRU))
	}
	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			// Replicated write capacity units (rWRU)
			{
				Name:            "Replicated write request unit (rWRU)",
				Unit:            "rWRU",
				UnitMultiplier:  schema.HourToMonthUnitMultiplier,
				MonthlyQuantity: quantity,
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
