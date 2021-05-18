package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMKeyVaultManagedHSMRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_key_vault_managed_hardware_security_module",
		RFunc: NewAzureKeyVaultManagedHSM,
	}
}

func NewAzureKeyVaultManagedHSM(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	location := d.Get("location").String()

	var costComponents []*schema.CostComponent

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           "HSM pools",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Key Vault"),
			ProductFamily: strPtr("Security"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Dedicated HSM")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Instance")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
