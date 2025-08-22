package azure

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

const (
	gbInBytes float64 = 1073741824
)

var (
	sqlEditionMapping = map[string]skuConfig{
		"GeneralPurpose": {
			sku:    "GP_Gen5_2",
			tier:   "General Purpose",
			family: "Compute Gen5",
			cores:  intPtr(2),
		},
		"BusinessCritical": {
			sku:    "BC_Gen5_2",
			tier:   "Business Critical",
			family: "Compute Gen5",
			cores:  intPtr(2),
		},
		"Hyperscale": {
			sku:    "HS_Gen5_2",
			tier:   "Hyperscale",
			family: "Compute Gen5",
			cores:  intPtr(2),
		},
		"Standard": {
			sku: "S0",
		},
		"Premium": {
			sku: "P1",
		},
		"DataWarehouse": {
			sku: "DW100c",
		},
		"Stretch": {
			sku: "DS100",
		},
	}
)

func getSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_sql_database",
		CoreRFunc: newSQLDatabase,
	}
}

func newSQLDatabase(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("location").String()

	config := skuConfig{
		sku:    "GP_Gen5_2",
		tier:   "General Purpose",
		family: "Compute Gen5",
		cores:  intPtr(2),
	}

	edition := d.Get("edition").String()
	if edition != "" {
		config = skuConfig{
			sku: edition,
		}

		if val, ok := sqlEditionMapping[edition]; ok {
			config = val
		}
	}

	sku := d.Get("requested_service_objective_name").String()
	if sku != "" {
		var err error
		config, err = parseSKU(d.Address, sku)
		if err != nil {
			logging.Logger.Warn().Msg(err.Error())
			return nil
		}
	}

	var maxSizeGB *float64
	maxBytes := d.Get("max_size_bytes").Float()
	if maxBytes > 0 {
		val := maxBytes / gbInBytes
		maxSizeGB = &val
	}

	var readReplicas *int64
	if d.Get("read_scale").Bool() {
		var i int64 = 1
		readReplicas = &i
	}

	r := &azure.SQLDatabase{
		Address:           d.Address,
		Region:            region,
		SKU:               config.sku,
		Tier:              config.tier,
		Family:            config.family,
		Cores:             config.cores,
		MaxSizeGB:         maxSizeGB,
		ReadReplicaCount:  readReplicas,
		ZoneRedundant:     d.Get("zone_redundant").Bool(),
		BackupStorageType: "Geo",
		IsDevTest:         d.ProjectMetadata["isProduction"] == "false",
	}
	return r
}

func parseSKU(address, sku string) (skuConfig, error) {
	if dtuMap.usesDTUUnits(sku) {
		return skuConfig{
			sku: sku,
		}, nil
	}

	return parseMSSQLSku(address, sku)
}
