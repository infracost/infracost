package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_db_instance",
		RFunc: NewDBInstance,
	}
}

func NewDBInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	deploymentOption := "Single-AZ"
	if d.Get("multi_az").Bool() {
		deploymentOption = "Multi-AZ"
	}

	instanceType := d.Get("instance_class").String()

	var databaseEngine *string
	switch d.Get("engine").String() {
	case "postgres":
		databaseEngine = strPtr("PostgreSQL")
	case "mysql":
		databaseEngine = strPtr("MySQL")
	case "mariadb":
		databaseEngine = strPtr("MariaDB")
	case "aurora", "aurora-mysql":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
	case "oracle-se", "oracle-se1", "oracle-se2", "oracle-ee":
		databaseEngine = strPtr("Oracle")
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		databaseEngine = strPtr("SQL Server")
	}

	var databaseEdition *string
	switch d.Get("engine").String() {
	case "oracle-se", "sqlserver-se":
		databaseEdition = strPtr("Standard")
	case "oracle-se1":
		databaseEdition = strPtr("Standard One")
	case "oracle-se2":
		databaseEdition = strPtr("Standard Two")
	case "oracle-ee", "sqlserver-ee":
		databaseEdition = strPtr("Enterprise")
	case "sqlserver-ex":
		databaseEdition = strPtr("Express")
	case "sqlserver-web":
		databaseEdition = strPtr("Web")
	}

	var licenseModel *string
	engineVal := d.Get("engine").String()
	if engineVal == "oracle-se1" || engineVal == "oracle-se2" || strings.HasPrefix(engineVal, "sqlserver-") {
		licenseModel = strPtr("License included")
	}
	if d.Get("license_model").String() == "bring-your-own-license" {
		licenseModel = strPtr("Bring your own license")
	}

	volumeType := "General Purpose"
	if d.Get("storage_type").Exists() {
		if d.Get("iops").Exists() && d.Get("iops").Type != gjson.Null {
			volumeType = "Provisioned IOPS"
		} else if d.Get("storage_type").String() == "standard" {
			volumeType = "Magnetic"
		} else if d.Get("storage_type").String() == "io1" {
			volumeType = "Provisioned IOPS"
		}
	}

	allocatedStorageVal := decimal.Zero
	if d.Get("allocated_storage").Exists() {
		allocatedStorageVal = decimal.NewFromFloat(d.Get("allocated_storage").Float())
	}

	iopsVal := decimal.Zero
	if d.Get("iops").Exists() {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	instanceAttributeFilters := []*schema.AttributeFilter{
		{Key: "instanceType", Value: strPtr(instanceType)},
		{Key: "deploymentOption", Value: strPtr(deploymentOption)},
		{Key: "databaseEngine", Value: databaseEngine},
	}
	if databaseEdition != nil {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "databaseEdition",
			Value: databaseEdition,
		})
	}
	if licenseModel != nil {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "licenseModel",
			Value: licenseModel,
		})
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           "Database instance",
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:       strPtr("aws"),
				Region:           strPtr(region),
				Service:          strPtr("AmazonRDS"),
				ProductFamily:    strPtr("Database Instance"),
				AttributeFilters: instanceAttributeFilters,
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name:            "Database storage",
			Unit:            "GB-months",
			MonthlyQuantity: &allocatedStorageVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeType", Value: strPtr(volumeType)},
					{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				},
			},
		},
	}

	if volumeType == "Provisioned IOPS" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Database storage IOPS",
			Unit:            "IOPS-months",
			MonthlyQuantity: &iopsVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Provisioned IOPS"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
