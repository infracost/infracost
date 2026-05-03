package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// StorageTable struct represents Azure Table Storage.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/storage/tables/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/storage/tables/#pricing
type StorageTable struct {
	Address                string
	Region                 string
	AccountReplicationType string
	HasCustomerManagedKey  bool

	MonthlyStorageGB            *float64 `infracost_usage:"monthly_storage_gb"`
	MonthlyBatchWriteOperations *int64   `infracost_usage:"monthly_batch_write_operations"`
	MonthlyWriteOperations      *int64   `infracost_usage:"monthly_write_operations"`
	MonthlyReadOperations       *int64   `infracost_usage:"monthly_read_operations"`
	MonthlyScanOperations       *int64   `infracost_usage:"monthly_scan_operations"`
	MonthlyListOperations       *int64   `infracost_usage:"monthly_list_operations"`
	MonthlyDeleteOperations     *int64   `infracost_usage:"monthly_delete_operations"`
}

func (r *StorageTable) CoreType() string {
	return "StorageTable"
}

func (r *StorageTable) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_batch_write_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_write_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_read_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_scan_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_list_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_delete_operations", DefaultValue: 0, ValueType: schema.Int64},
	}
}

func (r *StorageTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *StorageTable) BuildResource() *schema.Resource {
	if !r.isReplicationTypeSupported() {
		logging.Logger.Warn().Msgf("Skipping resource %s. Storage Tables don't support %s redundancy", r.Address, r.AccountReplicationType)
		return nil
	}

	costComponents := []*schema.CostComponent{
		r.dataStorageCostComponent(),
	}
	costComponents = append(costComponents, r.operationsCostComponent("Batch Write", r.MonthlyBatchWriteOperations))
	costComponents = append(costComponents, r.operationsCostComponent("Write", r.MonthlyWriteOperations))
	costComponents = append(costComponents, r.operationsCostComponent("Read", r.MonthlyReadOperations))
	costComponents = append(costComponents, r.operationsCostComponent("Scan", r.MonthlyScanOperations))
	costComponents = append(costComponents, r.operationsCostComponent("List", r.MonthlyListOperations))
	costComponents = append(costComponents, r.operationsCostComponent("Delete", r.MonthlyDeleteOperations))

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *StorageTable) isReplicationTypeSupported() bool {
	validReplicationTypes := []string{"LRS", "ZRS", "GRS", "RA-GRS", "GZRS", "RA-GZRS"}

	return contains(validReplicationTypes, strings.ToUpper(r.AccountReplicationType))
}

func (r *StorageTable) lookupPrefix() string {
	if r.HasCustomerManagedKey {
		return "Account Encrypted"
	}
	return "Standard"
}

func (r *StorageTable) dataStorageCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyStorageGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageGB))
	}

	replication := strings.ToUpper(r.AccountReplicationType)
	// Account Encrypted GZRS/RA-GZRS Data Stored meters carry the "Account Encrypted"
	// prefix in the meter name; the other replication types do not.
	meterName := fmt.Sprintf("%s Data Stored", replication)
	if r.HasCustomerManagedKey && (replication == "GZRS" || replication == "RA-GZRS") {
		meterName = fmt.Sprintf("Account Encrypted %s Data Stored", replication)
	}

	return &schema.CostComponent{
		Name:            "Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Tables")},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%s %s", r.lookupPrefix(), replication))},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *StorageTable) operationsCostComponent(operation string, usage *int64) *schema.CostComponent {
	var qty *decimal.Decimal
	if usage != nil {
		qty = decimalPtr(decimal.NewFromInt(*usage).Div(decimal.NewFromInt(10000)))
	}

	// RA-GZRS is a special case. There's no listed prices for the write operations, but they
	// use the same price as the read operations
	operationLookup := operation
	if !r.HasCustomerManagedKey && strings.EqualFold(r.AccountReplicationType, "ra-gzrs") && (operation == "Batch Write" || operation == "Write") {
		operationLookup = "Read"
	}

	if operationLookup == "Write" {
		// Exclude "Batch Write Operations" from this lookup. ValueRegex is
		// evaluated server-side by the pricing API (PCRE), so the lookbehind
		// is fine here even though Go's RE2 wouldn't compile it.
		operationLookup = "(?<!Batch\\s+)Write"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s operations", cases.Title(language.English).String(operation)),
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Tables")},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%s %s", r.lookupPrefix(), strings.ToUpper(r.AccountReplicationType)))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Operations$", operationLookup))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}