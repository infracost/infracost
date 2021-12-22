package azure

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

var (
	sqlTierMapping = map[string]string{
		"GP":   "General Purpose",
		"GP_S": "General Purpose - Serverless",
		"HS":   "Hyperscale",
		"BC":   "Business Critical",
	}

	sqlFamilyMapping = map[string]string{
		"Gen5": "Compute Gen5",
		"Gen4": "Compute Gen4",
		"M":    "Compute M Series",
	}

	dtuMap = dtuMapping{
		"free":  true,
		"basic": true,

		"s": true, // covers Standard, System editions
		"d": true, // covers DataWarehouse editions
		"p": true, // covers Premium editions
	}
)

type dtuMapping map[string]bool

func (d dtuMapping) usesDTUUnits(sku string) bool {
	sanitized := strings.ToLower(sku)
	if d[sanitized] {
		return true
	}

	return d[sanitized[0:1]]
}

func getAzureRMMSSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mssql_database",
		RFunc: newAzureRMMSSQLDatabase,
		ReferenceAttributes: []string{
			"server_id",
		},
	}
}

func newAzureRMMSSQLDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"server_id"})

	var sku string
	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}

	var maxSize *float64
	if d.Get("max_size_gb").Type != gjson.Null {
		val := d.Get("max_size_gb").Float()
		maxSize = &val
	}

	var replicaCount *int64
	if d.Get("read_replica_count").Exists() {
		val := d.Get("read_replica_count").Int()
		replicaCount = &val
	}

	licenceType := "LicenseIncluded"
	if d.Get("license_type").Exists() {
		licenceType = d.Get("license_type").String()
	}

	r := &azure.SQLDatabase{
		Address:          d.Address,
		Region:           region,
		SKU:              sku,
		LicenceType:      licenceType,
		MaxSizeGB:        maxSize,
		ReadReplicaCount: replicaCount,
		ZoneRedundant:    d.Get("zone_redundant").Bool(),
	}

	if !dtuMap.usesDTUUnits(sku) {
		c, err := parseMSSQLSku(d.Address, sku)
		if err != nil {
			log.Warnf(err.Error())
			return nil
		}

		r.Tier = c.tier
		r.Family = c.family
		r.Cores = c.cores
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}

type skuConfig struct {
	sku    string
	tier   string
	family string
	cores  *int64
}

func parseMSSQLSku(address, sku string) (skuConfig, error) {
	s := strings.Split(sku, "_")
	if len(s) < 3 {
		return skuConfig{}, fmt.Errorf("Unrecognized MSSQL SKU format for resource %s: %s", address, sku)
	}

	tierKey := strings.Join(s[0:len(s)-2], "_")
	tier, ok := sqlTierMapping[tierKey]
	if !ok {
		return skuConfig{}, fmt.Errorf("Invalid tier in MSSQL SKU for resource %s: %s", address, sku)
	}

	familyKey := s[len(s)-2]
	family, ok := sqlFamilyMapping[familyKey]
	if !ok {
		return skuConfig{}, fmt.Errorf("Invalid family in MSSQL SKU for resource %s: %s", address, sku)
	}

	cores, err := strconv.ParseInt(s[len(s)-1], 10, 64)
	if err != nil {
		return skuConfig{}, fmt.Errorf("Invalid core count in MSSQL SKU for resource %s: %s", address, sku)
	}

	return skuConfig{
		sku:    sku,
		tier:   tier,
		family: family,
		cores:  &cores,
	}, nil
}
