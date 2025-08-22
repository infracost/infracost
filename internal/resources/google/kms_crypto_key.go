package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type KMSCryptoKey struct {
	Address              string
	Region               string
	VersionTemplate      string
	Algorithm            string
	ProtectionLevel      string
	RotationPeriod       string
	KeyVersions          *int64 `infracost_usage:"key_versions"`
	MonthlyKeyOperations *int64 `infracost_usage:"monthly_key_operations"`
}

func (r *KMSCryptoKey) CoreType() string {
	return "KMSCryptoKey"
}

func (r *KMSCryptoKey) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "key_versions", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_key_operations", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *KMSCryptoKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KMSCryptoKey) BuildResource() *schema.Resource {
	region := r.Region

	algorithm := "GOOGLE_SYMMETRIC_ENCRYPTION"
	protectionLevel := "SOFTWARE"

	if r.VersionTemplate != "" {
		algorithm = r.Algorithm
		protectionLevel = r.ProtectionLevel
	}

	var monthlyKeys *decimal.Decimal
	if r.KeyVersions != nil {
		monthlyKeys = decimalPtr(decimal.NewFromInt(*r.KeyVersions))
	} else if r.RotationPeriod != "" {
		rotation, err := strconv.ParseFloat(strings.Split(r.RotationPeriod, "s")[0], 64)

		if err == nil {
			monthlyKeys = decimalPtr(decimal.NewFromFloat(2592000.0 / rotation))
		}
	}

	var monthlyKeyOperations *decimal.Decimal
	if r.MonthlyKeyOperations != nil {
		monthlyKeyOperations = decimalPtr(decimal.NewFromInt(*r.MonthlyKeyOperations))
	}

	var keyDescript = r.cryptoKeyDescription(algorithm, protectionLevel)
	var operationDesctipt = keyOperationsDescription(algorithm, protectionLevel)

	costComponents := []*schema.CostComponent{}

	if strings.ToLower(keyDescript) == "hsm rsa 3072" || strings.ToLower(keyDescript) == "hsm rsa 4096" || strings.ToLower(keyDescript) == "hsm ecdsa p-256" || strings.ToLower(keyDescript) == "hsm ecdsa p-384" {
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
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: firstTierQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Cloud Key Management Service (KMS)"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s/i", keyDescript))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr("2000"),
			},
			UsageBased: true,
		})

		if len(tiers) > 1 && tiers[1].GreaterThan(decimal.NewFromInt(0)) {

			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Key versions (over 2K)",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: &tiers[1],
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Cloud Key Management Service (KMS)"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s/i", keyDescript))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("2000"),
				},
				UsageBased: true,
			})
		}
	} else {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Key versions",
			Unit:            "months",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: monthlyKeys,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Cloud Key Management Service (KMS)"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s/i", keyDescript))},
				},
			},
			UsageBased: true,
		})
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Operations",
		Unit:            "10k operations",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: monthlyKeyOperations,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Cloud Key Management Service (KMS)"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s/i", operationDesctipt))},
			},
		},
		UsageBased: true,
	})

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *KMSCryptoKey) cryptoKeyDescription(algorithm string, protectionLevel string) string {
	protectionLevel = strings.ToLower(protectionLevel)
	switch protectionLevel {
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
		if strings.HasPrefix(strings.ToLower(algorithm), "rsa_sign_") {
			parts := strings.Split(algorithm, "_")
			if len(parts) > 3 {
				return "HSM RSA " + parts[3]
			}
		}
	}
	return ""
}

func keyOperationsDescription(algorithm string, protectionLevel string) string {
	protectionLevel = strings.ToLower(protectionLevel)
	switch protectionLevel {
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
		if strings.HasPrefix(strings.ToLower(algorithm), "rsa_sign_") {
			parts := strings.Split(algorithm, "_")
			if len(parts) > 3 {
				return "HSM cryptographic operations with a RSA   " + parts[3]
			}
		}
	}
	return ""
}
