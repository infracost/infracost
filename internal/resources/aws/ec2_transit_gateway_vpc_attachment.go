package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Ec2TransitGatewayVpcAttachment struct {
	Address                string
	Region                 string
	VPCRegion              string
	TransitGatewayRegion   string
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *Ec2TransitGatewayVpcAttachment) CoreType() string {
	return "Ec2TransitGatewayVpcAttachment"
}

func (r *Ec2TransitGatewayVpcAttachment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *Ec2TransitGatewayVpcAttachment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Ec2TransitGatewayVpcAttachment) BuildResource() *schema.Resource {
	region := r.Region

	if r.VPCRegion != "" {
		region = r.VPCRegion
	}

	if r.TransitGatewayRegion != "" {
		region = r.TransitGatewayRegion
	}

	var gbDataProcessed *decimal.Decimal

	if r.MonthlyDataProcessedGB != nil {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayAttachmentCostComponent(region, "TransitGatewayVPC"),
			transitGatewayDataProcessingCostComponent(region, "TransitGatewayVPC", gbDataProcessed),
		}, UsageSchema: r.UsageSchema(),
	}
}

func transitGatewayAttachmentCostComponent(region string, operation string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Transit gateway attachment",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonVPC"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Hours/")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
	}
}

func transitGatewayDataProcessingCostComponent(region string, operation string, gbDataProcessed *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: gbDataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonVPC"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Bytes/")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
		UsageBased: true,
	}
}
