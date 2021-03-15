package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type SQLInstanceDBType int

const (
	MySQL SQLInstanceDBType = iota
	PostgreSQL
	SQLServer
)

func GetSQLInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_sql_database_instance",
		RFunc:               NewSQLInstance,
		ReferenceAttributes: []string{},
	}
}

func NewSQLInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	name := d.Address
	tier := d.Get("settings.0").Get("tier").String()
	availabilityType := d.Get("settings.0").Get("availability_type").String()
	region := d.Get("region").String()
	dbVersion := d.Get("database_version").String()
	dbType := SQLInstanceDBVersionToDBType(dbVersion)
	return &schema.Resource{
		Name: name,
		CostComponents: []*schema.CostComponent{
			SharedSQLInstance(name, tier, availabilityType, dbType, region),
		},
	}
}

func SharedSQLInstance(name, tier, availabilityType string, dbType SQLInstanceDBType, region string) *schema.CostComponent {
	cost := &schema.CostComponent{
		Name: "Instance pricing",
	}
	resourceGroup := SQLInstanceTierToResourceGroup(tier)
	if resourceGroup == "" {
		log.Debugf("No tier resource group for sql instance %s", name)
		return cost
	}
	descriptionRegex := SQLInstanceAvDBTypeToDescriptionRegex(availabilityType, dbType)
	cost = &schema.CostComponent{
		Name:           "Instance pricing",
		Unit:           "seconds",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
				{Key: "description", ValueRegex: strPtr(descriptionRegex)},
			},
		},
	}
	return cost
}

func SQLInstanceDBVersionToDBType(dbVersion string) SQLInstanceDBType {
	if strings.Contains(dbVersion, "POSTGRES") {
		return PostgreSQL
	} else if strings.Contains(dbVersion, "MYSQL") {
		return MySQL
	} else if strings.Contains(dbVersion, "SQLSERVER") {
		return SQLServer
	} else {
		return MySQL
	}
}

func SQLInstanceTierToResourceGroup(tier string) string {
	data := map[string]string{
		"db-f1-micro": "SQLGen2InstancesF1Micro",
		"db-g1-small": "SQLGen2InstancesF1Small",
	}
	return data[tier]
}

func SQLInstanceAvDBTypeToDescriptionRegex(availabilityType string, dbType SQLInstanceDBType) string {
	dbTypeNames := map[SQLInstanceDBType]string{
		MySQL:      "MySQL",
		PostgreSQL: "PostgreSQL",
		SQLServer:  "SQL Server",
	}
	availabilityTypeNames := map[string]string{
		"REGIONAL": "Regional",
		"ZONAL":    "Zonal",
	}
	dbTypeString := dbTypeNames[dbType]
	availabilityTypeString := availabilityTypeNames[availabilityType]
	description := fmt.Sprintf("/%s: %s/", dbTypeString, availabilityTypeString)
	return description
}
