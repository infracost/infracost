package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
)

type SSMParameter struct {
	Address                *string
	Region                 *string
	Tier                   *string
	MonthlyApiInteractions *int64  `infracost_usage:"monthly_api_interactions"`
	ParameterStorageHrs    *int64  `infracost_usage:"parameter_storage_hrs"`
	ApiThroughputLimit     *string `infracost_usage:"api_throughput_limit"`
}

var SSMParameterUsageSchema = []*schema.UsageItem{{Key: "monthly_api_interactions", ValueType: schema.Int64, DefaultValue: 0}, {Key: "parameter_storage_hrs", ValueType: schema.Int64, DefaultValue: 0}, {Key: "api_throughput_limit", ValueType: schema.String, DefaultValue: "standard"}}

func (r *SSMParameter) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SSMParameter) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	storage := parameterStorageCostComponent(r)
	if storage != nil {
		costComponents = append(costComponents, storage)
	}
	apiThroughput := apiThroughputCostComponent(r)
	if apiThroughput != nil {
		costComponents = append(costComponents, apiThroughput)
	}
	if len(costComponents) == 0 {
		return &schema.Resource{
			Name:      *r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: SSMParameterUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: SSMParameterUsageSchema,
	}
}

func parameterStorageCostComponent(r *SSMParameter,) *schema.CostComponent {
	region := *r.Region

	tier := "Standard"
	if r.Tier != nil {
		tier = *r.Tier
	}
	if strings.ToLower(tier) == "standard" {

		return nil
	}

	parameterStorageHours := decimal.NewFromInt(730)
	if r != nil && r.ParameterStorageHrs != nil {
		parameterStorageHours = decimal.NewFromInt(*r.ParameterStorageHrs)
	}

	return &schema.CostComponent{
		Name:            "Parameter storage (advanced)",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &parameterStorageHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("AWS Systems Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/PS-Advanced-Param-Tier1/")},
			},
		},
	}
}

func apiThroughputCostComponent(r *SSMParameter,) *schema.CostComponent {
	region := *r.Region

	tier := "standard"
	if r.Tier != nil {
		tier = *r.Tier
	}
	if r != nil && r.ApiThroughputLimit != nil {
		tier = *r.ApiThroughputLimit
	}
	tier = strings.ToLower(tier)

	if tier == "standard" {

		return nil
	}
	if !(tier == "advanced" || tier == "higher") {
		log.Errorf("api_throughput_limit in %s must be one of: advanced, higher", *r.Address)
	}

	var monthlyAPIInteractions *decimal.Decimal
	if r != nil && r.MonthlyApiInteractions != nil {
		monthlyAPIInteractions = decimalPtr(decimal.NewFromInt(*r.MonthlyApiInteractions))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("API interactions (%s)", tier),
		Unit:            "10k interactions",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: monthlyAPIInteractions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("API Request"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/PS-Param-Processed-Tier2/")},
			},
		},
	}
}
