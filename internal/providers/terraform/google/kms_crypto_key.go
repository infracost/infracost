package google

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetKMSCryptoKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_kms_crypto_key",
		RFunc: NewKMSCryptoKey,
	}
}

func NewKMSCryptoKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	algorithm := d.Get("version_template.0.algorithm").String()
	protectionLevel := d.Get("version_template.0.protection_level").String()

	if algorithm == "EXTERNAL_SYMMETRIC_ENCRYPTION" {
		protectionLevel = "" // default value is SOFTWARE, and EXTERNAL isn't possible value. ???
	}

	var monthlyKeys *decimal.Decimal
	if u != nil && u.Get("monthly_key_versions").Exists() {
		monthlyKeys = decimalPtr(decimal.NewFromInt(u.Get("monthly_key_versions").Int()))
	}

	var monthlyKeyOperations *decimal.Decimal
	if u != nil && u.Get("monthly_key_operations").Exists() {
		monthlyKeyOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_key_operations").Int()))
	}

	var keyDescript = cryptoKeyDescription(algorithm, protectionLevel)
	var operationDesctipt = keyOperationsDescription(algorithm, protectionLevel)

	costComponents := []*schema.CostComponent{}

	if keyDescript == "HSM RSA 3072" || keyDescript == "HSM RSA 4096" || keyDescript == "HSM ECDSA P-256" || keyDescript == "HSM ECDSA P-384" {
		tierLimit := decimal.NewFromInt(2000)
		tierOneKeys := decimal.NewFromInt(0)
		tierTwoKeys := decimal.NewFromInt(0)

		if monthlyKeys != nil {
			if monthlyKeys.GreaterThan(tierLimit) {
				tierOneKeys = tierLimit
				tierTwoKeys = monthlyKeys.Sub(tierLimit)
			} else {
				tierOneKeys = *monthlyKeys
			}
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            keyDescript + " (first 2K)",
			Unit:            "keys",
			UnitMultiplier:  1,
			MonthlyQuantity: &tierOneKeys,
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

		if tierTwoKeys.GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            keyDescript,
				Unit:            "keys",
				UnitMultiplier:  1,
				MonthlyQuantity: &tierTwoKeys,
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
			Name:            keyDescript,
			Unit:            "keys",
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
		Name:            operationDesctipt,
		Unit:            "operations",
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
	switch protectionLevel {
	case "SOFTWARE":
		if algorithm == "GOOGLE_SYMMETRIC_ENCRYPTION" {
			return "Software symmetric"
		}
		return "Software asymmetric"
	case "HSM":
		if algorithm == "GOOGLE_SYMMETRIC_ENCRYPTION" {
			return "HSM symmetric"
		}
		if algorithm == "EC_SIGN_P256_SHA256" {
			return "HSM ECDSA P-256"
		}
		if algorithm == "EC_SIGN_P384_SHA384" {
			return "HSM ECDSA P-384"
		}
		rsaType := strings.Split(algorithm, "_")[3]
		return "HSM RSA " + rsaType
	}
	if algorithm == "EXTERNAL_SYMMETRIC_ENCRYPTION" {
		return "external symmetric"
	}
	return ""
}

func keyOperationsDescription(algorithm string, protectionLevel string) string {
	switch protectionLevel {
	case "SOFTWARE":
		if algorithm == "GOOGLE_SYMMETRIC_ENCRYPTION" {
			return "Cryptographic operations with a software symmetric"
		}
		return "Software asymmetric cryptographic"
	case "HSM":
		if algorithm == "GOOGLE_SYMMETRIC_ENCRYPTION" {
			return "HSM symmetric cryptographic"
		}
		if algorithm == "EC_SIGN_P256_SHA256" {
			return "HSM cryptographic operations with an ECDSA P-256"
		}
		if algorithm == "EC_SIGN_P384_SHA384" {
			return "HSM cryptographic operations with an ECDSA P-384"
		}
		rsaType := strings.Split(algorithm, "_")[3]
		return "HSM cryptographic operations with a RSA " + rsaType
	}
	if algorithm == "EXTERNAL_SYMMETRIC_ENCRYPTION" {
		return "External symmetric cryptographic"
	}
	return ""
}
