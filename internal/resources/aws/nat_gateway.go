package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type NATGatewayArguments struct {
	Address string `json:"address,omitempty"`
	Region  string `json:"region,omitempty"`

	MonthlyDataProcessedGB *float64 `json:"monthlyDataProcessedGB,omitempty"`
}

func (args *NATGatewayArguments) PopulateUsage(u *schema.UsageData) {
	if u != nil {
		args.MonthlyDataProcessedGB = u.GetFloat("monthly_data_processed_gb")
	}
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
		Name:        args.Address,
		UsageSchema: NATGatewayUsageSchema,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "NAT gateway",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(args.Region),
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
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataProcessed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(args.Region),
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
