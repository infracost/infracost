package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// StorageShare struct represents an Azure Files Storage Shares
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/storage/files/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/storage/files/#pricing
type StorageShare struct {
	Address                string
	Region                 string
	AccountReplicationType string
	AccessTier             string
	Quota                  int64

	// "usage" args
	MonthlyStorageGB        *float64 `infracost_usage:"storage_gb"`
	MonthlyReadOperations   *int64   `infracost_usage:"monthly_read_operations"`
	MonthlyWriteOperations  *int64   `infracost_usage:"monthly_write_operations"`
	MonthlyListOperations   *int64   `infracost_usage:"monthly_list_operations"`
	MonthlyOtherOperations  *int64   `infracost_usage:"monthly_other_operations"`
	MonthlyDataRetrievalGB  *float64 `infracost_usage:"monthly_data_retrieval_gb"`
	SnapshotsStorageGB      *float64 `infracost_usage:"snapshots_storage_gb"`
	MetadataAtRestStorageGB *float64 `infracost_usage:"metadata_at_rest_storage_gb"`
}

// CoreType returns the name of this resource type
func (r *StorageShare) CoreType() string {
	return "StorageShare"
}

// UsageSchema defines a list which represents the usage schema of StorageShare.
func (r *StorageShare) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_read_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_write_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_list_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_other_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_retrieval_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "snapshots_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "metadata_at_rest_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the StorageShare.
// It uses the `infracost_usage` struct tags to populate data into the StorageShare.
func (r *StorageShare) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid StorageShare struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *StorageShare) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.dataStorageCostComponent(),
	}

	costComponents = append(costComponents, r.snapshotCostComponents()...)
	costComponents = append(costComponents, r.metadataCostComponents()...)
	costComponents = append(costComponents, r.readOperationsCostComponents()...)
	costComponents = append(costComponents, r.writeOperationsCostComponents()...)
	costComponents = append(costComponents, r.listOperationsCostComponents()...)
	costComponents = append(costComponents, r.otherOperationsCostComponents()...)
	costComponents = append(costComponents, r.dataRetrievalCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *StorageShare) productName() string {
	if r.accessTier() == "Premium" {
		return "Premium Files"
	}

	return "Files v2"
}

func (r *StorageShare) accessTier() string {
	return map[string]string{
		"hot":                  "Hot",
		"cool":                 "Cool",
		"transactionoptimized": "Standard",
		"premium":              "Premium",
	}[strings.ToLower(r.AccessTier)]
}

func (r *StorageShare) dataStorageCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal

	if r.accessTier() == "Premium" {
		qty = decimalPtr(decimal.NewFromInt(r.Quota))
	}

	if r.MonthlyStorageGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageGB))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))
	meterName := "Data Stored"
	if r.accessTier() == "Premium" {
		meterName = "Provisioned"
	}

	return &schema.CostComponent{
		Name:            "Data at rest",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *StorageShare) snapshotCostComponents() []*schema.CostComponent {
	var qty *decimal.Decimal
	if r.SnapshotsStorageGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.SnapshotsStorageGB))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))
	meterName := "Data Stored"
	if r.accessTier() == "Premium" {
		meterName = "Snapshots"
	}

	return []*schema.CostComponent{{
		Name:            "Snapshots",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}}
}

func (r *StorageShare) metadataCostComponents() []*schema.CostComponent {
	if contains([]string{"Premium", "Standard"}, r.accessTier()) {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MetadataAtRestStorageGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MetadataAtRestStorageGB))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))

	return []*schema.CostComponent{{
		Name:            "Metadata at rest",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", ValueRegex: regexPtr("Metadata$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}}
}

func (r *StorageShare) readOperationsCostComponents() []*schema.CostComponent {
	if r.accessTier() == "Premium" {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MonthlyReadOperations != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyReadOperations).Div(decimal.NewFromInt(10000)))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))

	return []*schema.CostComponent{{
		Name:            "Read operations",
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", ValueRegex: regexPtr("Read Operations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}}
}

func (r *StorageShare) writeOperationsCostComponents() []*schema.CostComponent {
	if r.accessTier() == "Premium" {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MonthlyWriteOperations != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyWriteOperations).Div(decimal.NewFromInt(10000)))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))

	return []*schema.CostComponent{{
		Name:            "Write operations",
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", ValueRegex: regexPtr("Write Operations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}}
}

func (r *StorageShare) listOperationsCostComponents() []*schema.CostComponent {
	if r.accessTier() == "Premium" {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MonthlyListOperations != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyListOperations).Div(decimal.NewFromInt(10000)))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))

	return []*schema.CostComponent{{
		Name:            "List operations",
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(skuName)},

				{Key: "meterName", ValueRegex: regexPtr("List Operations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}}
}

func (r *StorageShare) otherOperationsCostComponents() []*schema.CostComponent {
	if r.accessTier() == "Premium" {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MonthlyOtherOperations != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyOtherOperations).Div(decimal.NewFromInt(10000)))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))
	meterName := "Other Operations"
	if r.accessTier() == "Standard" {
		meterName = "Protocol Operations"
	}

	return []*schema.CostComponent{
		{
			Name:            "Other operations",
			Unit:            "10k operations",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: qty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(r.productName())},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption:   strPtr("Consumption"),
				StartUsageAmount: strPtr("0"),
			},
			UsageBased: true,
		}}
}

func (r *StorageShare) dataRetrievalCostComponents() []*schema.CostComponent {
	if contains([]string{"Premium", "Standard", "Hot"}, r.accessTier()) || strings.ToUpper(r.AccountReplicationType) == "GZRS" {
		return []*schema.CostComponent{}
	}

	var qty *decimal.Decimal
	if r.MonthlyDataRetrievalGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataRetrievalGB))
	}

	skuName := fmt.Sprintf("%s %s", r.accessTier(), strings.ToUpper(r.AccountReplicationType))

	return []*schema.CostComponent{
		{
			Name:            "Data retrieval",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: qty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(r.productName())},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Data Retrieval$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption:   strPtr("Consumption"),
				StartUsageAmount: strPtr("0"),
			},
			UsageBased: true,
		}}
}
