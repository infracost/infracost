package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getExpressRouteGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_express_route_gateway",
		RFunc: newExpressRouteGateway,
	}
}

func newExpressRouteGateway(d *schema.ResourceData) schema.CoreResource {
	e := &azure.ExpressRouteGateway{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		ScaleUnits: d.Get("scaleUnits").Int(),
	}

	return e
}
