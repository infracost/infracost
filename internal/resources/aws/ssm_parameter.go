package aws

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type SSMParameter struct {
	Address                string
	Tier                   string
	Region                 string
	ParameterStorageHrs    *int64  `infracost_usage:"parameter_storage_hrs"`
	APIThroughputLimit     *string `infracost_usage:"api_throughput_limit"`
	MonthlyAPIInteractions *int64  `infracost_usage:"monthly_api_interactions"`
}

func (r *SSMParameter) CoreType() string {
	return "SSMParameter"
}

func (r *SSMParameter) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "parameter_storage_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "api_throughput_limit", ValueType: schema.String, DefaultValue: "standard"},
		{Key: "monthly_api_interactions", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SSMParameter) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SSMParameter) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	throughputLimit := ""

	if r.APIThroughputLimit != nil {
		throughputLimit = strings.ToLower(*r.APIThroughputLimit)

		if throughputLimit != "standard" && throughputLimit != "advanced" && throughputLimit != "higher" {
			logging.Logger.Warn().Msgf("Skipping resource %s. Unrecognized api_throughput_limit %s, expecting standard, advanced or higher", r.Address, *r.APIThroughputLimit)
			return nil
		}
	}

	if throughputLimit == "" {
		throughputLimit = r.tierValue()
	}

	if r.tierValue() != "standard" {
		costComponents = append(costComponents, r.parameterStorageCostComponent())
		costComponents = append(costComponents, r.apiThroughputCostComponent(throughputLimit))
	}

	if len(costComponents) == 0 {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *SSMParameter) tierValue() string {
	if r.Tier == "" {
		return "standard"
	}

	return strings.ToLower(r.Tier)
}

func (r *SSMParameter) parameterStorageCostComponent() *schema.CostComponent {
	parameterStorageHours := decimal.NewFromInt(730)
	if r.ParameterStorageHrs != nil {
		parameterStorageHours = decimal.NewFromInt(*r.ParameterStorageHrs)
	}

	return &schema.CostComponent{
		Name:            "Parameter storage (advanced)",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &parameterStorageHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("AWS Systems Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/PS-Advanced-Param-Tier1/")},
			},
		},
		UsageBased: true,
	}
}

func (r *SSMParameter) apiThroughputCostComponent(throughputLimit string) *schema.CostComponent {
	var monthlyAPIInteractions *decimal.Decimal
	if r.MonthlyAPIInteractions != nil {
		monthlyAPIInteractions = decimalPtr(decimal.NewFromInt(*r.MonthlyAPIInteractions))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("API interactions (%s)", throughputLimit),
		Unit:            "10k interactions",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: monthlyAPIInteractions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("API Request"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/PS-Param-Processed-Tier2/")},
			},
		},
		UsageBased: true,
	}
}
