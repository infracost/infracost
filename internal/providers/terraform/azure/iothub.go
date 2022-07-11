package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func getIoTHubRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_iothub",
		RFunc: newIoTHub,
	}
}

func getIoTHubDPSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_iothub_dps",
		RFunc: newIoTHubDPS,
	}
}

func newIoTHub(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	sku := d.Get("sku.0.name").String()
	capacity := d.Get("sku.0.capacity").Int()

	r := &azure.IoTHub{
		Address:  d.Address,
		Region:   region,
		Sku:      sku,
		Capacity: capacity,
		DPS:      false,
	}

	r.PopulateUsage(u)

	return r.BuildResource()
}

func newIoTHubDPS(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	sku := d.Get("sku.0.name").String()
	capacity := d.Get("sku.0.capacity").Int()

	if u != nil && u.Get("monthly_operations").Type != gjson.Null {
		r := &azure.IoTHub{
			Address:  d.Address,
			Region:   region,
			Sku:      sku,
			Capacity: capacity,
			DPS:      true,
		}

		r.PopulateUsage(u)

		return r.BuildResource()
	}

	log.Warnf("Skipping resource %s. Could not find a way to get its cost components from the resource or usage file.", d.Address)
	return nil
}
