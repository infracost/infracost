package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetDXGatewayAssociationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dx_gateway_association",
		RFunc: NewDXGatewayAssociation,
	}
}

func NewDXGatewayAssociation(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataProcessed *decimal.Decimal

	if u != nil && u.Get("monthly_gb_data_processed.0.value").Exists(){
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_processed.0.value").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayDataProcessingCostComponent(region, "TransitGatewayDirectConnect", gbDataProcessed),
			transitGatewayAttachmentCostComponent(region, "TransitGatewayDirectConnect"),
		},
	}
}