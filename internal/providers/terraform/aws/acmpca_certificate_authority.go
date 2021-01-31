package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

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

	certificateTierLimits := []int{1000, 10000, 10001}

	monthlyCertificatesRequests := decimal.Zero

	if u != nil && u.Get("monthly_certificate_requests").Exists() {
		monthlyCertificatesRequests = decimal.NewFromInt(u.Get("monthly_certificate_requests").Int())
	}

	privateCertificateTier := usage.CalculateTierRequests(monthlyCertificatesRequests, certificateTierLimits)

	tierOne := privateCertificateTier["1"]
	tierTwo := privateCertificateTier["2"]
	tierThree := privateCertificateTier["3"]

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

	if tierOne.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (1 - 1000)", "0", tierOne))
	}

	if tierTwo.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (1001 - 10000)", "1000", tierTwo))
	}

	if tierThree.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, certificateCostComponent(region, "Certificate requests (> 10000)", "10000", tierThree))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func certificateCostComponent(region string, displayName string, usageTier string, monthlyQuantity decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "requests",
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
