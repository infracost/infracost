package azure

import (
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPostgreSQLFlexibleServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "dbforpostgresql.azure.upbound.io/FlexibleServer",
		CoreRFunc: newPostgreSQLFlexibleServer,
	}
}

func newPostgreSQLFlexibleServer(d *schema.ResourceData) schema.CoreResource {
	forProvider := d.Get("forProvider")

	region := lookupRegion(d, []string{})
	
	sku := forProvider.Get("skuName").String()
	storage := forProvider.Get("storageMb").Int()

	var tier, size, version string

	s := strings.Split(sku, "_")
	if len(s) < 3 || len(s) > 4 {
		logging.Logger.Warn().Msgf("Unrecognised PostgreSQL Flexible Server SKU format for resource %s: %s", d.Address, sku)
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
		logging.Logger.Warn().Msgf("Unrecognised PostgreSQL Flexible Server tier prefix for resource %s: %s", d.Address, sku)
		return nil
	}

	if tier != "b" {
		coreRegex := regexp.MustCompile(`(\d+)`)
		match := coreRegex.FindStringSubmatch(size)
		if len(match) < 1 {
			logging.Logger.Warn().Msgf("Unrecognised PostgreSQL Flexible Server size for resource %s: %s", d.Address, sku)
			return nil
		}
	}

	r := &azure.PostgreSQLFlexibleServer{
		Address:         d.Address,
		Region:          region,
		SKU:             sku,
		Tier:            tier,
		InstanceType:    size,
		InstanceVersion: version,
		Storage:         storage,
	}
	return r
}
