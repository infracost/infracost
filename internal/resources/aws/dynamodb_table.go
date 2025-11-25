package aws

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
)

type DynamoDBTable struct {
	// "required" args that can't really be missing.
	Address        string
	Region         string
	Name           string
	BillingMode    string
	ReplicaRegions []string

	// "optional" args, that may be empty depending on the resource config
	WriteCapacity *int64
	ReadCapacity  *int64

	AppAutoscalingTarget       []*AppAutoscalingTarget
	PointInTimeRecoveryEnabled bool

	// "usage" args
	MonthlyWriteRequestUnits       *int64 `infracost_usage:"monthly_write_request_units"`
	MonthlyReadRequestUnits        *int64 `infracost_usage:"monthly_read_request_units"`
	StorageGB                      *int64 `infracost_usage:"storage_gb"`
	PitrBackupStorageGB            *int64 `infracost_usage:"pitr_backup_storage_gb"`
	OnDemandBackupStorageGB        *int64 `infracost_usage:"on_demand_backup_storage_gb"`
	MonthlyDataRestoredGB          *int64 `infracost_usage:"monthly_data_restored_gb"`
	MonthlyStreamsReadRequestUnits *int64 `infracost_usage:"monthly_streams_read_request_units"`
}

func (a *DynamoDBTable) CoreType() string {
	return "DynamoDBTable"
}

func (a *DynamoDBTable) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_write_request_units", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "pitr_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "on_demand_backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_restored_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_streams_read_request_units", DefaultValue: 0, ValueType: schema.Int64},
	}
}

func (a *DynamoDBTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *DynamoDBTable) BuildResource() *schema.Resource {

	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	if a.BillingMode == "PROVISIONED" {
		var wcuAutoscaling, rcuAutoscaling bool
		wcu := a.WriteCapacity
		rcu := a.ReadCapacity

		for _, target := range a.AppAutoscalingTarget {
			switch target.ScalableDimension {
			case "dynamodb:table:WriteCapacityUnits":
				wcuAutoscaling = true
				if target.Capacity != nil {
					wcu = target.Capacity
				} else {
					wcu = &target.MinCapacity
				}
			case "dynamodb:table:ReadCapacityUnits":
				rcuAutoscaling = true
				if target.Capacity != nil {
					rcu = target.Capacity
				} else {
					rcu = &target.MinCapacity
				}
			}
		}
		// Write capacity units (WCU)
		costComponents = append(costComponents, a.wcuCostComponent(a.Region, wcu, wcuAutoscaling))
		// Read capacity units (RCU)
		costComponents = append(costComponents, a.rcuCostComponent(a.Region, rcu, rcuAutoscaling))
	}

	// Infracost usage data

	if a.BillingMode == "PAY_PER_REQUEST" {
		// Write request units (WRU)
		costComponents = append(costComponents, a.wruCostComponent(a.Region, a.MonthlyWriteRequestUnits))
		// Read request units (RRU)
		costComponents = append(costComponents, a.rruCostComponent(a.Region, a.MonthlyReadRequestUnits))
	}

	// Data storage
	costComponents = append(costComponents, a.dataStorageCostComponent(a.Region, a.StorageGB))
	// Continuous backups (PITR)
	if a.PointInTimeRecoveryEnabled {
		costComponents = append(costComponents, a.continuousBackupCostComponent(a.Region, a.PitrBackupStorageGB))
	}

	// OnDemand backups
	costComponents = append(costComponents, a.onDemandBackupCostComponent(a.Region, a.OnDemandBackupStorageGB))
	// Restoring tables
	costComponents = append(costComponents, a.restoreCostComponent(a.Region, a.MonthlyDataRestoredGB))

	// Stream reads
	costComponents = append(costComponents, a.streamCostComponent(a.Region, a.MonthlyStreamsReadRequestUnits))

	// Global tables (replica)
	subResources = append(subResources, a.globalTables(a.BillingMode, a.ReplicaRegions, a.WriteCapacity, a.MonthlyWriteRequestUnits)...)

	estimate := func(ctx context.Context, values map[string]any) error {
		storageB, err := aws.DynamoDBGetStorageBytes(ctx, a.Region, a.Name)
		if err != nil {
			return err
		}
		values["storage_gb"] = asGiB(storageB)

		if a.BillingMode == "PAY_PER_REQUEST" {
			reads, err := aws.DynamoDBGetRRU(ctx, a.Region, a.Name)
			if err != nil {
				return err
			}
			writes, err := aws.DynamoDBGetWRU(ctx, a.Region, a.Name)
			if err != nil {
				return err
			}
			values["monthly_read_request_units"] = ceil64(reads)
			values["monthly_write_request_units"] = ceil64(writes)
		}
		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		EstimateUsage:  estimate,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func (a *DynamoDBTable) wcuCostComponent(region string, provisionedWCU *int64, autoscaling bool) *schema.CostComponent {
	name := "Write capacity unit (WCU)"
	if autoscaling {
		name = "Write capacity unit (WCU, autoscaling)"
	}

	var quantity *decimal.Decimal
	if provisionedWCU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*provisionedWCU))
	}
	return &schema.CostComponent{
		Name:           name,
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
			DescriptionRegex: regexPtr("^(?!.*\\(free tier\\)).*$"),
		},
		UsageBased: autoscaling,
	}
}

func (a *DynamoDBTable) rcuCostComponent(region string, provisionedRCU *int64, autoscaling bool) *schema.CostComponent {
	name := "Read capacity unit (RCU)"
	if autoscaling {
		name = "Read capacity unit (RCU, autoscaling)"
	}

	var quantity *decimal.Decimal
	if provisionedRCU != nil {
		quantity = decimalPtr(decimal.NewFromInt(*provisionedRCU))
	}
	return &schema.CostComponent{
		Name:           name,
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
			DescriptionRegex: regexPtr("^(?!.*\\(free tier\\)).*$"),
		},
		UsageBased: autoscaling,
	}
}

func (a *DynamoDBTable) globalTables(billingMode string, replicaRegions []string, writeCapacity *int64, monthlyWRU *int64) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for _, region := range replicaRegions {
		name := fmt.Sprintf("Global table (%s)", region)
		if billingMode == "PROVISIONED" {
			resources = append(resources, a.newProvisionedDynamoDBGlobalTable(name, region, writeCapacity))
		} else if billingMode == "PAY_PER_REQUEST" {
			resources = append(resources, a.newOnDemandDynamoDBGlobalTable(name, region, monthlyWRU))
		}
	}

	return resources
}

func (a *DynamoDBTable) newProvisionedDynamoDBGlobalTable(name string, region string, provisionedWCU *int64) *schema.Resource {
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
					DescriptionRegex: regexPtr("^(?!.*\\(free tier\\)).*$"),
				},
			},
		},
	}
}

func (a *DynamoDBTable) newOnDemandDynamoDBGlobalTable(name string, region string, monthlyWRU *int64) *schema.Resource {
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

func (a *DynamoDBTable) wruCostComponent(region string, monthlyWRU *int64) *schema.CostComponent {
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
		UsageBased: true,
	}
}

func (a *DynamoDBTable) rruCostComponent(region string, monthlyRRU *int64) *schema.CostComponent {
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
		UsageBased: true,
	}
}

func (a *DynamoDBTable) dataStorageCostComponent(region string, storageGB *int64) *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: regexPtr("(?<!IA-)TimedStorage-ByteHrs$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			DescriptionRegex: regexPtr("^(?!.*\\$0.00 per GB-Month).*$"),
		},
		UsageBased: true,
	}
}

func (a *DynamoDBTable) continuousBackupCostComponent(region string, pitrBackupStorageGB *int64) *schema.CostComponent {
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
		UsageBased: true,
	}
}

func (a *DynamoDBTable) onDemandBackupCostComponent(region string, onDemandBackupStorageGB *int64) *schema.CostComponent {
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
		UsageBased: true,
	}
}

func (a *DynamoDBTable) restoreCostComponent(region string, monthlyDataRestoredGB *int64) *schema.CostComponent {
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
		UsageBased: true,
	}
}

func (a *DynamoDBTable) streamCostComponent(region string, monthlyStreamsRRU *int64) *schema.CostComponent {
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
			DescriptionRegex: regexPtr("^(?!.*\\(free tier\\)).*$"),
		},
		UsageBased: true,
	}
}
