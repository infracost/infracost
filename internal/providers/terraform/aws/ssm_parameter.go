package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
)

func GetSSMParameterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_parameter",
		RFunc: NewSSMParameter,
	}
}

func NewSSMParameter(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	storage := parameterStorageCostComponent(d, u)
	if storage != nil {
		costComponents = append(costComponents, storage)
	}
	apiThroughput := apiThroughputCostComponent(d, u)
	if apiThroughput != nil {
		costComponents = append(costComponents, apiThroughput)
	}
	if len(costComponents) == 0 {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func parameterStorageCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()

	tier := "Standard"
	if d.Get("tier").Exists() {
		tier = d.Get("tier").String()
	}
	if tier == "Standard" {
		// Standard is free
		return nil
	}

	parameterStorageHours := decimal.NewFromInt(730)
	if u != nil && u.Get("parameter_storage_hours").Exists() {
		parameterStorageHours = decimal.NewFromInt(u.Get("parameter_storage_hours").Int())
	}

	return &schema.CostComponent{
		Name:            "Parameter storage (advanced)",
		Unit:            "hours",
		UnitMultiplier:  1,
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

func apiThroughputCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()

	tier := "standard"
	if d.Get("tier").Exists() {
		tier = d.Get("tier").String()
	}
	if u != nil && u.Get("api_throughput_limit").Exists() {
		tier = u.Get("api_throughput_limit").String()
	}
	tier = strings.ToLower(tier)

	if tier == "standard" {
		// Standard is free
		return nil
	}
	if !(tier == "advanced" || tier == "higher") {
		log.Errorf("api_throughput_limit in %s must be one of: advanced, higher", d.Address)
	}

	var monthlyAPIInteractions *decimal.Decimal
	if u != nil && u.Get("monthly_api_interactions").Exists() {
		monthlyAPIInteractions = decimalPtr(decimal.NewFromInt(u.Get("monthly_api_interactions").Int()))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("API interactions (%s)", tier),
		Unit:            "interactions",
		UnitMultiplier:  10000,
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
