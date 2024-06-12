package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getApplicationGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_application_gateway",
		CoreRFunc: NewApplicationGateway,
	}
}

func NewApplicationGateway(d *schema.ResourceData) schema.CoreResource {
	var autoscalingMinCapacity *int64
	if d.Get("autoscale_configuration.0.min_capacity").Exists() {
		autoscalingMinCapacity = intPtr(d.Get("autoscale_configuration.0.min_capacity").Int())
	}

	r := &azure.ApplicationGateway{
		Address:                d.Address,
		SKUName:                d.Get("sku.0.name").String(),
		SKUCapacity:            d.Get("sku.0.capacity").Int(),
		AutoscalingMinCapacity: autoscalingMinCapacity,
		Region:                 d.Region,
	}

	return r
}
