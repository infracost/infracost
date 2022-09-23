package azure

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMySQLFlexibleServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mysql_flexible_server",
		RFunc: newMySQLFlexibleServer,
	}
}

func newMySQLFlexibleServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	sku := d.Get("sku_name").String()
	storage := d.GetInt64OrDefault("storage.0.size_gb", 0)
	iops := d.GetInt64OrDefault("storage.0.iops", 0)

	var tier, size, version string

	s := strings.Split(sku, "_")
	if len(s) < 3 || len(s) > 4 {
		log.Warnf("Unrecognised MySQL Flexible Server SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	if len(s) > 2 {
		tier = strings.ToLower(s[0])
		size = s[2]
	}

	if len(s) > 3 {
		version = s[3]
	}

	supportedTiers := []string{"b", "gp", "mo"}
	if !contains(supportedTiers, tier) {
		log.Warnf("Unrecognised MySQL Flexible Server tier prefix for resource %s: %s", d.Address, sku)
		return nil
	}

	if tier != "b" {
		coreRegex := regexp.MustCompile(`(\d+)`)
		match := coreRegex.FindStringSubmatch(size)
		if len(match) < 1 {
			log.Warnf("Unrecognised MySQL Flexible Server size for resource %s: %s", d.Address, sku)
			return nil
		}
	}

	r := &azure.MySQLFlexibleServer{
		Address:         d.Address,
		Region:          region,
		SKU:             sku,
		Tier:            tier,
		InstanceType:    size,
		InstanceVersion: version,
		Storage:         storage,
		IOPS:            iops,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
