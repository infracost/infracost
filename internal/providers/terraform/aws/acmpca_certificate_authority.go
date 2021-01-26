package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetACMPCACertificateAuthorityRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_acmpca_certificate_authority",
		RFunc: NewACMPCACertificateAuthority,
	}
}

func NewACMPCACertificateAuthority(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var privateCertificateTier = map[string]decimal.Decimal{
		"tierOne":   decimal.Zero,
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
	}

	monthlyCertificatesCreated := decimal.Zero

	if u != nil && u.Get("monthly_certificates_created").Exists() {
		monthlyCertificatesCreated = decimal.NewFromInt(u.Get("monthly_certificates_created").Int())
	}

	certificateCreationQuantities := calculateCertificateRequests(monthlyCertificatesCreated, privateCertificateTier)

	tierOne := certificateCreationQuantities["tierOne"]
	tierTwo := certificateCreationQuantities["tierTwo"]
	tierThree := certificateCreationQuantities["tierThree"]

	costComponents := []*schema.CostComponent{
		{
			Name:            "Private certificate authority",
			Unit:            "authority",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSCertificateManager"),
				ProductFamily: strPtr("AWS Certificate Manager"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/PaidPrivateCA/")},
				},
			},
		},
	}

	if privateCertificateTier["tierOne"].GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Certificates (1 - 1000)",
			Unit:            "certificates",
			UnitMultiplier:  1,
			MonthlyQuantity: &tierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSCertificateManager"),
				ProductFamily: strPtr("AWS Certificate Manager"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/PrivateCertificatesIssued/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr("1000"),
			},
		})
	}

	if privateCertificateTier["tierTwo"].GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Certificates (1,0001 - 10,000)",
			Unit:            "certificates",
			UnitMultiplier:  1,
			MonthlyQuantity: &tierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSCertificateManager"),
				ProductFamily: strPtr("AWS Certificate Manager"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/PrivateCertificatesIssued/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr("10000"),
			},
		})
	}

	if privateCertificateTier["tierThree"].GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Certificates (10,001 and above)",
			Unit:            "certificates",
			UnitMultiplier:  1,
			MonthlyQuantity: &tierThree,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSCertificateManager"),
				ProductFamily: strPtr("AWS Certificate Manager"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/PrivateCertificatesIssued/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("10000"),
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func calculateCertificateRequests(privateCertificateCount decimal.Decimal, pricingTiers map[string]decimal.Decimal) map[string]decimal.Decimal {
	certificateTierOneLimit := decimal.NewFromInt(1000)
	certificateTierTwoLimit := decimal.NewFromInt(10000)
	certificateTierThreeLimit := decimal.NewFromInt(10001)

	if privateCertificateCount.GreaterThan(certificateTierOneLimit) {
		pricingTiers["tierOne"] = certificateTierOneLimit
	} else {
		pricingTiers["tierOne"] = privateCertificateCount
		return pricingTiers
	}

	if privateCertificateCount.GreaterThan(certificateTierTwoLimit) {
		pricingTiers["tierTwo"] = certificateTierTwoLimit
	} else {
		pricingTiers["tierTwo"] = privateCertificateCount.Sub(certificateTierOneLimit)
	}

	if privateCertificateCount.GreaterThan(certificateTierThreeLimit) {
		pricingTiers["tierThree"] = certificateTierThreeLimit
	} else {
		pricingTiers["tierThree"] = privateCertificateCount.Sub(certificateTierTwoLimit.Add(certificateTierOneLimit))
		return pricingTiers
	}

	return pricingTiers
}
