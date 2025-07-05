package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type RecaptchaEnterpriseKey struct {
	Address             string
	Region              string
	MonthlyAssessments  *int64 `infracost_usage:"monthly_assessments"`
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
	var monthlyAssessments int64
	if r.MonthlyAssessments != nil {
		monthlyAssessments = *r.MonthlyAssessments
	}

	costComponents := []*schema.CostComponent{}

	switch {
	case monthlyAssessments <= 10000:
		// Free tier
	case monthlyAssessments <= 100000:
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "reCAPTCHA Enterprise assessments (10K–100K tier)",
			Unit:            "monthly flat fee",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("reCAPTCHA Enterprise"),
				ProductFamily: strPtr("Application Services"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr("Assessment")},
				},
			},
		})
	default:
		extraAssessments := monthlyAssessments - 100000
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "reCAPTCHA Enterprise assessments (>100K tier)",
			Unit:            "1k assessments",
			UnitMultiplier:  decimal.NewFromInt(1000),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(extraAssessments)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("reCAPTCHA Enterprise"),
				ProductFamily: strPtr("Application Services"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr("Assessment")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
