package azure

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

var (
	sqlTierMapping = map[string]string{
		"gp":   "General Purpose",
		"gp_s": "General Purpose - Serverless",
		"hs":   "Hyperscale",
		"bc":   "Business Critical",
	}

	sqlFamilyMapping = map[string]string{
		"gen5": "Compute Gen5",
		"gen4": "Compute Gen4",
		"m":    "Compute M Series",
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

	if sanitized == "" {
		return false
	}

	return d[sanitized[0:1]]
}

func getMSSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_mssql_database",
		CoreRFunc: newAzureRMMSSQLDatabase,
		ReferenceAttributes: []string{
			"server_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"server_id"})
		},
	}
}

func newAzureRMMSSQLDatabase(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	sku := d.GetStringOrDefault("sku_name", "GP_S_Gen5_2")

	var maxSize *float64
	if !d.IsEmpty("max_size_gb") {
		val := d.Get("max_size_gb").Float()
		maxSize = &val
	}

	var replicaCount *int64
	if !d.IsEmpty("read_replica_count") {
		val := d.Get("read_replica_count").Int()
		replicaCount = &val
	}

	licenseType := d.GetStringOrDefault("license_type", "LicenseIncluded")
	storageAccountType := d.GetStringOrDefault("storage_account_type", "Geo")

	r := &azure.SQLDatabase{
		Address:           d.Address,
		Region:            region,
		SKU:               sku,
		LicenseType:       licenseType,
		MaxSizeGB:         maxSize,
		ReadReplicaCount:  replicaCount,
		ZoneRedundant:     d.Get("zone_redundant").Bool(),
		BackupStorageType: storageAccountType,
		IsDevTest:         d.ProjectMetadata["isProduction"] == "false",
	}

	if strings.ToLower(sku) == "elasticpool" || !d.IsEmpty("elastic_pool_id") {
		r.IsElasticPool = true
	} else if !dtuMap.usesDTUUnits(sku) {
		c, err := parseMSSQLSku(d.Address, sku)
		if err != nil {
			logging.Logger.Warn().Msg(err.Error())
			return nil
		}

		r.Tier = c.tier
		r.Family = c.family
		r.Cores = c.cores
	}

	return r
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

	tierKey := strings.ToLower(strings.Join(s[0:len(s)-2], "_"))
	tier, ok := sqlTierMapping[tierKey]
	if !ok {
		return skuConfig{}, fmt.Errorf("Invalid tier in MSSQL SKU for resource %s: %s", address, sku)
	}

	familyKey := strings.ToLower(s[len(s)-2])
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
