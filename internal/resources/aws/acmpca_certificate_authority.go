package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type ACMPCACertificateAuthority struct {
	Address         string
	Region          string
	UsageMode       string
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

func (r *ACMPCACertificateAuthority) CoreType() string {
	return "ACMPCACertificateAuthority"
}

func (r *ACMPCACertificateAuthority) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *ACMPCACertificateAuthority) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ACMPCACertificateAuthority) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.certificateAuthorityCostComponent(),
	}

	if r.MonthlyRequests != nil {
		monthlyCertificatesRequests := decimal.NewFromInt(*r.MonthlyRequests)

		if r.shortLived() {
			costComponents = append(costComponents, r.certificateCostComponent("Certificates (short-lived)", "0", &monthlyCertificatesRequests))
		} else {
			certificateTierLimits := []int{1000, 9000}
			certificateTiers := usage.CalculateTierBuckets(monthlyCertificatesRequests, certificateTierLimits)

			if certificateTiers[0].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.certificateCostComponent("Certificates (first 1K)", "0", &certificateTiers[0]))
			}

			if certificateTiers[1].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.certificateCostComponent("Certificates (next 9K)", "1000", &certificateTiers[1]))
			}

			if certificateTiers[2].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.certificateCostComponent("Certificates (over 10K)", "10000", &certificateTiers[2]))
			}
		}
	} else {
		var unknown *decimal.Decimal
		if r.shortLived() {
			costComponents = append(costComponents, r.certificateCostComponent("Certificates (short-lived)", "0", unknown))
		} else {
			costComponents = append(costComponents, r.certificateCostComponent("Certificates (first 1K)", "0", unknown))
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ACMPCACertificateAuthority) shortLived() bool {
	return strings.ToLower(r.UsageMode) == "short_lived_certificate"
}

func (r *ACMPCACertificateAuthority) certificateAuthorityCostComponent() *schema.CostComponent {
	name := "Private certificate authority"
	regex := "/-PaidPrivateCA/"
	if r.shortLived() {
		name = "Private certificate authority (short-lived certificate mode)"
		regex = "/-ShortLivedCertificatePrivateCA/"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSCertificateManager"),
			ProductFamily: strPtr("AWS Certificate Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: &regex},
			},
		},
	}
}

func (r *ACMPCACertificateAuthority) certificateCostComponent(displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	regex := "/-PrivateCertificatesIssued/"
	if r.shortLived() {
		regex = "/-ShortLivedCertificatesIssued/"
	}

	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "requests",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSCertificateManager"),
			ProductFamily: strPtr("AWS Certificate Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: &regex},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
