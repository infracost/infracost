package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type RecaptchaEnterpriseKey struct {
	Address            string
	Region             string
	MonthlyAssessments *int64 `infracost_usage:"monthly_assessments"`
}

func (r *RecaptchaEnterpriseKey) CoreType() string {
	return "RecaptchaEnterpriseKey"
}

func (r *RecaptchaEnterpriseKey) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_assessments", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *RecaptchaEnterpriseKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RecaptchaEnterpriseKey) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	monthlyAssessments := int64(0)
	if r.MonthlyAssessments != nil {
		monthlyAssessments = *r.MonthlyAssessments
	}

	tierLimits := []int{10000, 100000}
	tierQuantities := usage.CalculateTierBuckets(decimal.NewFromInt(monthlyAssessments), tierLimits)


	// Tier: 10K–100K
	if tierQuantities[1].GreaterThan(decimal.Zero) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "reCAPTCHA Enterprise assessments (10K–100K tier)",
			Unit:            "100K assessments",
			UnitMultiplier:  decimal.NewFromInt(100000),
			MonthlyQuantity: &tierQuantities[1],
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("reCAPTCHA Enterprise"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr("Recaptcha")},
					{Key: "description", Value: strPtr("reCAPTCHA Create Assessment Requests")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("10000"),
			},
		})
	}

	// Tier: >100K
	if tierQuantities[2].GreaterThan(decimal.Zero) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "reCAPTCHA Enterprise assessments (>100K tier)",
			Unit:            "1k assessments",
			UnitMultiplier:  decimal.NewFromInt(1000),
			MonthlyQuantity: &tierQuantities[2],
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("reCAPTCHA Enterprise"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr("Recaptcha")},
					{Key: "description", Value: strPtr("reCAPTCHA Create Assessment Requests")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("100000"),
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
