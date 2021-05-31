package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func GetAzureRMKeyVaultCertificateRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_key_vault_certificate",
		RFunc: NewAzureKeyVaultCertificate,
		ReferenceAttributes: []string{
			"key_vault_id",
		},
	}
}

func NewAzureKeyVaultCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var location, skuName string
	keyVault := d.References("key_vault_id")
	location = keyVault[0].Get("location").String()

	if location == "" {
		log.Warnf("Skipping resource %s. Could not find its 'location' property.", d.Address)
		return nil
	}

	var costComponents []*schema.CostComponent
	skuName = strings.Title(keyVault[0].Get("sku_name").String())

	var certificateRenewals, certificateOperations *decimal.Decimal
	if u != nil && u.Get("monthly_certificate_renewal_requests").Exists() {
		certificateRenewals = decimalPtr(decimal.NewFromInt(u.Get("monthly_certificate_renewal_requests").Int()))
	}
	costComponents = append(costComponents, vaultKeysCostComponent(
		"Certificate renewals",
		location,
		"requests",
		skuName,
		"Certificate Renewal Request",
		"0",
		certificateRenewals,
		1))

	if u != nil && u.Get("monthly_certificate_other_operations").Exists() {
		certificateOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_certificate_other_operations").Int()))
	}
	costComponents = append(costComponents, vaultKeysCostComponent(
		"Certificate operations",
		location,
		"10K transactions",
		skuName,
		"Operations",
		"0",
		certificateOperations,
		10000))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
