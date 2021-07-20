package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type NATGatewayArguments struct {
	Address *string `json:"address,omitempty"`
	Region  *string `json:"region,omitempty" infracost_usage:"region,us-east1,Region where gateway is located,infracost"`

	MonthlyDataProcessedGB *float64 `json:"monthlyDataProcessedGB,omitempty" infracost_usage:"monthly_data_processed_gb,0,Monthly data processed by the NAT Gateway in GB,infracost,terraform"`
}

func (args *NATGatewayArguments) PopulateArgs(u *schema.UsageData) {
	address := strPtr("nat_gateway") // TODO: Better default value?
	region := strPtr("")             // TODO: FIXME: A default value?
	monthlyDataProcessedGB := floatPtr(0.0)
	if u != nil {
		if u.Get("region").Exists() {
			usageRegion := u.Get("region").String()
			region = &usageRegion
		}
		if u.Get("monthly_data_processed_gb").Exists() {
			monthlyDataProcessedGB = u.GetFloat("monthly_data_processed_gb")
		}
	}
	args.Address = address
	args.Region = region
	args.MonthlyDataProcessedGB = monthlyDataProcessedGB
}

var NATGatewayUsageSchema = []*schema.UsageSchemaItem{
	{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
}

func NewNATGateway(args *NATGatewayArguments) *schema.Resource {
	var gbDataProcessed *decimal.Decimal
	if args.MonthlyDataProcessedGB != nil {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(*args.MonthlyDataProcessedGB))
	}

	return &schema.Resource{
		Name: *args.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "NAT gateway",
				Unit:           "hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        args.Region,
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Hours/")},
					},
				},
			},
			{
				Name:            "Data processed",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbDataProcessed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        args.Region,
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Bytes/")},
					},
				},
			},
		},
	}
}
