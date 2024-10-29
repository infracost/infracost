package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DXGatewayAssociation struct {
	Address                 string
	Region                  string
	AssociatedGatewayRegion string
	MonthlyDataProcessedGB  *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *DXGatewayAssociation) CoreType() string {
	return "DXGatewayAssociation"
}

func (r *DXGatewayAssociation) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *DXGatewayAssociation) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DXGatewayAssociation) BuildResource() *schema.Resource {
	region := r.Region

	if r.AssociatedGatewayRegion != "" {
		region = r.AssociatedGatewayRegion
	}

	var gbDataProcessed *decimal.Decimal

	if r.MonthlyDataProcessedGB != nil {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayDataProcessingCostComponent(region, "TransitGatewayDirectConnect", gbDataProcessed),
			transitGatewayAttachmentCostComponent(region, "TransitGatewayDirectConnect"),
		}, UsageSchema: r.UsageSchema(),
	}
}
