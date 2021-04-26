package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func GetAzureMSSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mssql_database",
		RFunc: NewAzureMSSQLDatabase,
	}
}

func NewAzureMSSQLDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	region := d.Get("location").String()
	serviceName := "SQL Database"

	var sku, computeType, tier, family, cores string
	var zoneRedundant bool
	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
		if s := strings.Split(sku, "_"); len(s) == 4 {
			if s[1] == "S" {
				computeType = "Serverless"
				tier = s[0]
				family = s[2]
				cores = s[3]
			}
		} else if s := strings.Split(sku, "_"); len(s) == 3 {
			computeType = "Provisioned"
			tier = s[0]
			family = s[1]
			cores = s[2]
		} else {
			log.Warnf("Unrecognised MSSQL SKU format for resource %s: %s", d.Address, sku)
			return nil
		}
	}

	tierName := map[string]string{
		"BC": "Business Critical",
		"GP": "General Purpose",
		"HS": "Hyperscale",
	}[tier]

	if d.Get("zone_redundant").Type != gjson.Null {
		zoneRedundant = d.Get("zone_redundant").Bool()
	}

	skuName := cores + " vCore"
	if zoneRedundant {
		skuName += " Zone Redundancy"
	}

	productNameRegex := "/" + tierName + " -"
	if computeType == "Serverless" {
		productNameRegex += " Serverless -"
	}
	productNameRegex += " Compute " + family + "/"

	fmt.Println(productNameRegex)
	fmt.Println(skuName)
	fmt.Println("Location:", region)

	costComponents = append(costComponents, databaseComputeInstance(region, serviceName, sku, productNameRegex, skuName))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
