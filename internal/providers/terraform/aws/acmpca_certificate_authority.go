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

	monthlyCertificatesRequests := decimal.Zero

	if u != nil && u.Get("monthly_certificate_requests").Exists() {
		monthlyCertificatesRequests = decimal.NewFromInt(u.Get("monthly_certificate_requests").Int())
	}

	certificateTierQuantities := calculateCertificateRequests(monthlyCertificatesRequests, privateCertificateTier)

	tierOne := certificateTierQuantities["tierOne"]
	tierTwo := certificateTierQuantities["tierTwo"]
	tierThree := certificateTierQuantities["tierThree"]

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
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (1 - 1000)", "0", tierOne))
	}

	if privateCertificateTier["tierTwo"].GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (1001 - 10000)", "1000", tierTwo))
	}

	if privateCertificateTier["tierThree"].GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (> 10000)", "10000", tierThree))
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
		return pricingTiers
	}

	if privateCertificateCount.GreaterThan(certificateTierThreeLimit) {
		pricingTiers["tierThree"] = privateCertificateCount.Sub(certificateTierTwoLimit.Add(certificateTierOneLimit))
	}

	return pricingTiers
}

func certificateCostComponent(region string, displayName string, usageTier string, monthlyQuantity decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "certificates",
		UnitMultiplier:  1,
		MonthlyQuantity: &monthlyQuantity,
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
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
