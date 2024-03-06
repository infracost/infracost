package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getAPIGatewayRestAPIRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_api_gateway_rest_api",
		CoreRFunc: NewAPIGatewayRestAPI,
	}
}
func NewAPIGatewayRestAPI(d *schema.ResourceData) schema.CoreResource {
	r := &aws.APIGatewayRestAPI{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
