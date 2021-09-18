package aws

import (
	"context"
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
	"github.com/shopspring/decimal"
)

type DynamoDbTableArguments struct {
	// "required" args that can't really be missing.
	Address        string
	Region         string
	Name           string
	BillingMode    string
	ReplicaRegions []string

	// "optional" args, that may be empty depending on the resource config
	WriteCapacity *int64
	ReadCapacity  *int64

	// "usage" args
	MonthlyWriteRequestUnits       *int64 `infracost_usage:"monthly_write_request_units"`
	MonthlyReadRequestUnits        *int64 `infracost_usage:"monthly_read_request_units"`
	StorageGB                      *int64 `infracost_usage:"storage_gb"`
	PitrBackupStorageGB            *int64 `infracost_usage:"pitr_backup_storage_gb"`
	OnDemandBackupStorageGB        *int64 `infracost_usage:"on_demand_backup_storage_gb"`
	MonthlyDataRestoredGB          *int64 `infracost_usage:"monthly_data_restored_gb"`
	MonthlyStreamsReadRequestUnits *int64 `infracost_usage:"monthly_streams_read_request_units"`
}

func (args *DynamoDbTableArguments) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(args, u)
}

var DynamoDBTableUsageSchema = []*schema.UsageSchemaItem{
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

	estimate := func(ctx context.Context, values map[string]interface{}) error {
		storageB, err := aws.DynamoDBGetStorageBytes(ctx, args.Region, args.Name)
		if err != nil {
			return err
		}
		values["storage_gb"] = asGiB(storageB)

		if args.BillingMode == "PAY_PER_REQUEST" {
			reads, err := aws.DynamoDBGetRRU(ctx, args.Region, args.Name)
			if err != nil {
				return err
			}
			writes, err := aws.DynamoDBGetWRU(ctx, args.Region, args.Name)
			if err != nil {
				return err
			}
			values["monthly_read_request_units"] = reads
			values["monthly_write_request_units"] = writes
		}
		return nil
	}

	return &schema.Resource{
		Name:           args.Address,
		UsageSchema:    DynamoDBTableUsageSchema,
		EstimateUsage:  estimate,
		CostComponents: costComponents,
		SubResources:   subResources,
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
