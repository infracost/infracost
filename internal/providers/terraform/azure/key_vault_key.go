package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMKeyVaultKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_key_vault_key",
		RFunc: NewAzureRMKeyVaultKey,
		ReferenceAttributes: []string{
			"key_vault_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"key_vault_id"})
		},
	}
}

func NewAzureRMKeyVaultKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var skuName, keyType, keySize, meterName string
	keyVault := d.References("key_vault_id")
	if len(keyVault) > 0 {
		skuName = cases.Title(language.English).String(keyVault[0].Get("sku_name").String())
	} else {
		logging.Logger.Warn().Msgf("Skipping resource %s. Could not find its 'sku_name' property on key_vault_id.", d.Address)
		return nil
	}

	keyType = d.Get("key_type").String()

	if d.Get("key_size").Type != gjson.Null {
		keySize = d.Get("key_size").String()
	}

	var costComponents []*schema.CostComponent

	unit := "10K transactions"

	var secretsTransactions *decimal.Decimal
	if u != nil && u.Get("monthly_secrets_operations").Exists() {
		secretsTransactions = decimalPtr(decimal.NewFromInt(u.Get("monthly_secrets_operations").Int()))
	}
	meterName = "Operations"
	costComponents = append(costComponents, vaultKeysCostComponent("Secrets operations", region, unit, skuName, meterName, "0", secretsTransactions, 10000))

	var keyRotationRenewals *decimal.Decimal
	if u != nil && u.Get("monthly_key_rotation_renewals").Exists() {
		keyRotationRenewals = decimalPtr(decimal.NewFromInt(u.Get("monthly_key_rotation_renewals").Int()))
	}
	meterName = "Secret Renewal"
	costComponents = append(costComponents, vaultKeysCostComponent("Storage key rotations", region, "renewals", skuName, meterName, "0", keyRotationRenewals, 1))

	if !strings.HasSuffix(keyType, "HSM") {
		var softwareProtectedTransactions *decimal.Decimal
		if u != nil && u.Get("monthly_protected_keys_operations").Exists() {
			softwareProtectedTransactions = decimalPtr(decimal.NewFromInt(u.Get("monthly_protected_keys_operations").Int()))
		}

		name := "Software-protected keys"
		meterName = "Operations"

		if keyType == "RSA" && keySize == "2048" {
			costComponents = append(costComponents, vaultKeysCostComponent(name, region, unit, skuName, meterName, "0", softwareProtectedTransactions, 10000))
		} else {
			meterName = "Advanced Key Operations"
			costComponents = append(costComponents, vaultKeysCostComponent(name, region, unit, skuName, meterName, "0", softwareProtectedTransactions, 10000))
		}
	}

	if strings.HasSuffix(keyType, "HSM") && strings.ToLower(skuName) == "premium" {
		var protectedKeys, hsmProtectedTransactions *decimal.Decimal

		name := "HSM-protected keys"
		keyUnit := "months"

		if u != nil && u.Get("hsm_protected_keys").Exists() {
			protectedKeys = decimalPtr(decimal.NewFromInt(u.Get("hsm_protected_keys").Int()))

			if keyType == "RSA-HSM" && keySize == "2048" {
				meterName = "Premium HSM-protected RSA 2048-bit key"
				costComponents = append(costComponents, vaultKeysCostComponent(name, region, keyUnit, skuName, meterName, "0", protectedKeys, 1))
			} else {
				meterName = "Premium HSM-protected Advanced Key"

				tierLimits := []int{250, 1250, 2500}
				keysQuantities := usage.CalculateTierBuckets(*protectedKeys, tierLimits)

				costComponents = append(costComponents, vaultKeysCostComponent(name+" (first 250)", region, keyUnit, skuName, meterName, "0", &keysQuantities[0], 1))
				if keysQuantities[1].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, vaultKeysCostComponent(name+" (next 1250)", region, keyUnit, skuName, meterName, "250", &keysQuantities[1], 1))
				}
				if keysQuantities[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, vaultKeysCostComponent(name+" (next 2500)", region, keyUnit, skuName, meterName, "1500", &keysQuantities[2], 1))
				}
				if keysQuantities[3].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, vaultKeysCostComponent(name+" (over 4000)", region, keyUnit, skuName, meterName, "4000", &keysQuantities[3], 1))
				}
			}
		} else {
			var unknown *decimal.Decimal
			costComponents = append(costComponents, vaultKeysCostComponent(name, region, keyUnit, skuName, meterName, "0", unknown, 1))
		}

		if u != nil && u.Get("monthly_protected_keys_operations").Exists() {
			hsmProtectedTransactions = decimalPtr(decimal.NewFromInt(u.Get("monthly_protected_keys_operations").Int()))

			if keyType == "RSA" && keySize == "2048" {
				meterName = "Operations"
				costComponents = append(costComponents, vaultKeysCostComponent(name, region, unit, skuName, meterName, "0", hsmProtectedTransactions, 10000))
			} else {
				meterName = "Advanced Key Operations"
				costComponents = append(costComponents, vaultKeysCostComponent(name, region, unit, skuName, meterName, "0", hsmProtectedTransactions, 10000))
			}
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vaultKeysCostComponent(name, region, unit, skuName, meterName, startUsage string, quantity *decimal.Decimal, multi int) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(multi))))
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Key Vault"),
			ProductFamily: strPtr("Security"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Key Vault")},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
