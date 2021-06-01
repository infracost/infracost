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
		RFunc: NewAzureRMKeyVaultCertificate,
		ReferenceAttributes: []string{
			"key_vault_id",
		},
	}
}

func NewAzureRMKeyVaultCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var skuName string
	keyVault := d.References("key_vault_id")
	if len(keyVault) > 0 {
		region = keyVault[0].Get("location").String()
	} else {
		log.Warnf("Using %s for resource %s as its 'location' property could not be found.", region, d.Address)
	}

	var costComponents []*schema.CostComponent
	skuName = strings.Title(keyVault[0].Get("sku_name").String())

	var certificateRenewals, certificateOperations *decimal.Decimal
	if u != nil && u.Get("monthly_certificate_renewal_requests").Exists() {
		certificateRenewals = decimalPtr(decimal.NewFromInt(u.Get("monthly_certificate_renewal_requests").Int()))
	}
	costComponents = append(costComponents, vaultKeysCostComponent(
		"Certificate renewals",
		region,
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
		region,
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
