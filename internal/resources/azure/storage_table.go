package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// StorageTable struct represents Azure Table Storage.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/storage/tables/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/storage/tables/#pricing
//
type StorageTable struct {
	Address                string
	Region                 string
	AccountKind            string
	AccountReplicationType string

	MonthlyStorageGB        *float64 `infracost_usage:"monthly_storage_gb"`
	MonthlyClass1Operations *int64   `infracost_usage:"monthly_class_1_operations"`
	MonthlyClass2Operations *int64   `infracost_usage:"monthly_class_2_operations"`
}

func (r *StorageTable) CoreType() string {
	return "StorageTable"
}

func (r *StorageTable) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_class_1_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_class_2_operations", DefaultValue: 0, ValueType: schema.Int64},
	}
}

func (r *StorageTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *StorageTable) BuildResource() *schema.Resource {
	if !r.isAccountKindSupported() {
		logging.Logger.Warn().Msgf("Skipping resource %s. Storage Tables don't support %s accounts", r.Address, r.AccountKind)
		return nil
	}

	if !r.isReplicationTypeSupported() {
		logging.Logger.Warn().Msgf("Skipping resource %s. Storage Tables don't support %s redundancy", r.Address, r.AccountReplicationType)
		return nil
	}

	costComponents := []*schema.CostComponent{
		r.dataStorageCostComponent(),
	}
	costComponents = append(costComponents, r.operationsCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *StorageTable) isAccountKindSupported() bool {
	return r.isStorageV1() || r.isStorageV2()
}

func (r *StorageTable) isReplicationTypeSupported() bool {
	var validReplicationTypes []string

	switch {
	case r.isStorageV1():
		validReplicationTypes = []string{"LRS", "GRS", "RA-GRS"}
	case r.isStorageV2():
		validReplicationTypes = []string{"LRS", "ZRS", "GRS", "RA-GRS", "GZRS", "RA-GZRS"}
	}

	if validReplicationTypes != nil {
		return contains(validReplicationTypes, strings.ToUpper(r.AccountReplicationType))
	}

	return true
}

func (r *StorageTable) isStorageV1() bool {
	return strings.EqualFold(r.AccountKind, "storage")
}

func (r *StorageTable) isStorageV2() bool {
	return strings.EqualFold(r.AccountKind, "storagev2")
}

func (r *StorageTable) productName() string {
	if r.isStorageV1() {
		return "Tables"
	}
	return "Tables v2"
}

func (r *StorageTable) dataStorageCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyStorageGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageGB))
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
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("Standard %s", strings.ToUpper(r.AccountReplicationType)))},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Data Stored", strings.ToUpper(r.AccountReplicationType)))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *StorageTable) operationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !contains([]string{"GZRS", "RA-GZRS"}, strings.ToUpper(r.AccountReplicationType)) {
		var class1Qty *decimal.Decimal
		if r.MonthlyClass1Operations != nil {
			class1Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyClass1Operations).Div(decimal.NewFromInt(10000)))
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Class 1 operations",
			Unit:            "10k operations",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: class1Qty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Storage"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr(r.productName())},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("Standard %s", strings.ToUpper(r.AccountReplicationType)))},
					{Key: "meterName", ValueRegex: regexPtr("Class 1 Operations$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption:   strPtr("Consumption"),
				StartUsageAmount: strPtr("0"),
			},
			UsageBased: true,
		})
	}

	var class2Qty *decimal.Decimal
	if r.MonthlyClass2Operations != nil {
		class2Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyClass2Operations).Div(decimal.NewFromInt(10000)))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Class 2 operations",
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: class2Qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(r.productName())},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("Standard %s", strings.ToUpper(r.AccountReplicationType)))},
				{Key: "meterName", ValueRegex: regexPtr("Class 2 Operations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	})

	return costComponents
} 