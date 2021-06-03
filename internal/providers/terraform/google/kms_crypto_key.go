package google

import (
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKMSCryptoKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_kms_crypto_key",
		RFunc: NewKMSCryptoKey,
	}
}

func NewKMSCryptoKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	algorithm := "GOOGLE_SYMMETRIC_ENCRYPTION"
	protectionLevel := "SOFTWARE"

	if d.Get("version_template").Type != gjson.Null {
		algorithm = d.Get("version_template.0.algorithm").String()
		protectionLevel = d.Get("version_template.0.protection_level").String()
	}

	var monthlyKeys *decimal.Decimal
	if u != nil && u.Get("key_versions").Exists() {
		monthlyKeys = decimalPtr(decimal.NewFromInt(u.Get("key_versions").Int()))
	} else if d.Get("rotation_period").Exists() {
		rotationPeriod := (d.Get("rotation_period").String())
		rotation, err := strconv.ParseFloat(strings.Split(rotationPeriod, "s")[0], 64)

		if err == nil {
			monthlyKeys = decimalPtr(decimal.NewFromFloat(2592000.0 / rotation))
		}
	}

	var monthlyKeyOperations *decimal.Decimal
	if u != nil && u.Get("monthly_key_operations").Exists() {
		monthlyKeyOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_key_operations").Int()))
	}

	var keyDescript = cryptoKeyDescription(algorithm, protectionLevel)
	var operationDesctipt = keyOperationsDescription(algorithm, protectionLevel)

	costComponents := []*schema.CostComponent{}

	if strings.ToLower(keyDescript) == "hsm rsa 3072" || strings.ToLower(keyDescript) == "hsm rsa 4096" || strings.ToLower(keyDescript) == "hsm ecdsa P-256" || strings.ToLower(keyDescript) == "hsm ecdsa p-384" {
		tierLimits := []int{2000}
		var firstTierQty *decimal.Decimal
		var tiers []decimal.Decimal
		if monthlyKeys != nil {
			tiers = usage.CalculateTierBuckets(*monthlyKeys, tierLimits)
			firstTierQty = &tiers[0]
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Key versions (first 2K)",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: firstTierQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Cloud Key Management Service (KMS)"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: strPtr("/" + keyDescript + "/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr("2000"),
			},
		})

		if tiers != nil && tiers[1].GreaterThan(decimal.NewFromInt(0)) {

			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Key versions (over 2K)",
				Unit:            "months",
				UnitMultiplier:  1,
				MonthlyQuantity: &tiers[1],
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Cloud Key Management Service (KMS)"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/" + keyDescript + "/")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("2000"),
				},
			})
		}
	} else {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Key versions",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: monthlyKeys,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Cloud Key Management Service (KMS)"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: strPtr("/" + keyDescript + "/")},
				},
			},
		})
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Operations",
		Unit:            "10k operations",
		UnitMultiplier:  10000,
		MonthlyQuantity: monthlyKeyOperations,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Cloud Key Management Service (KMS)"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/" + operationDesctipt + "/")},
			},
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func cryptoKeyDescription(algorithm string, protectionLevel string) string {
	switch strings.ToLower(protectionLevel) {
	case "software":
		if strings.ToLower(algorithm) == "google_symmetric_encryption" {
			return "Active software symmetric key versions"
		}
		return "Software asymmetric"
	case "hsm":
		if strings.ToLower(algorithm) == "google_symmetric_encryption" {
			return "HSM symmetric"
		}
		if strings.ToLower(algorithm) == "ec_sign_p256_sha256" {
			return "HSM ECDSA P-256"
		}
		if strings.ToLower(algorithm) == "ec_sign_p384_sha384" {
			return "HSM ECDSA P-384"
		}
		rsaType := strings.Split(algorithm, "_")[3]
		return "HSM RSA " + rsaType
	}
	return ""
}

func keyOperationsDescription(algorithm string, protectionLevel string) string {
	switch strings.ToLower(protectionLevel) {
	case "software":
		if strings.ToLower(algorithm) == "google_symmetric_encryption" {
			return "Cryptographic operations with a software symmetric"
		}
		return "Software asymmetric cryptographic"
	case "hsm":
		if strings.ToLower(algorithm) == "google_symmetric_encryption" {
			return "HSM symmetric cryptographic"
		}
		if strings.ToLower(algorithm) == "ec_sign_p256_sha256" {
			return "HSM cryptographic operations with an ECDSA P-256"
		}
		if strings.ToLower(algorithm) == "ec_sign_p384_sha384" {
			return "HSM cryptographic operations with an ECDSA P-384"
		}
		rsaType := strings.Split(algorithm, "_")[3]
		return "HSM cryptographic operations with a RSA " + rsaType
	}
	return ""
}
