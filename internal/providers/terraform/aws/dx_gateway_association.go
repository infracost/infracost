package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDXGatewayAssociationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_dx_gateway_association",
		CoreRFunc:           NewDXGatewayAssociation,
		ReferenceAttributes: []string{"associated_gateway_id"},
	}
}
func NewDXGatewayAssociation(d *schema.ResourceData) schema.CoreResource {
	r := &aws.DXGatewayAssociation{Address: d.Address, Region: d.Get("region").String()}

	// Try to get the region from the associated gateway
	assocGateway := d.References("associated_gateway_id")
	if len(assocGateway) > 0 {
		r.AssociatedGatewayRegion = assocGateway[0].Get("region").String()
	}
	return r
}
