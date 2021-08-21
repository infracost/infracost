package aws

import (
	"fmt"
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

func NewDBInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	deploymentOption := "Single-AZ"
	if d.Get("multi_az").Bool() {
		deploymentOption = "Multi-AZ"
	}

	instanceType := d.Get("instance_class").String()

	var monthlyIORequests *decimal.Decimal
	if u != nil && u.Get("monthly_standard_io_requests").Exists() {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_standard_io_requests").Int()))
	}

	var databaseEngine *string
	switch strings.ToLower(d.Get("engine").String()) {
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
	switch strings.ToLower(d.Get("engine").String()) {
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
	engineVal := strings.ToLower(d.Get("engine").String())
	if engineVal == "oracle-se1" || engineVal == "oracle-se2" || strings.HasPrefix(engineVal, "sqlserver-") {
		licenseModel = strPtr("License included")
	}
	if strings.ToLower(d.Get("license_model").String()) == "bring-your-own-license" {
		licenseModel = strPtr("Bring your own license")
	}

	iopsVal := decimal.Zero
	if d.Get("iops").Exists() && d.Get("iops").Type != gjson.Null {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	allocatedStorageVal := decimal.NewFromInt(20)
	if d.Get("allocated_storage").Exists() && d.Get("allocated_storage").Type != gjson.Null {
		allocatedStorageVal = decimal.NewFromFloat(d.Get("allocated_storage").Float())
	}

	volumeType := "General Purpose"
	storageName := "Storage (general purpose SSD, gp2)"
	if d.Get("storage_type").Exists() {
		if strings.ToLower(d.Get("storage_type").String()) == "io1" || iopsVal.GreaterThan(decimal.Zero) {
			volumeType = "Provisioned IOPS"
			storageName = "Storage (provisioned IOPS SSD, io1)"
			if iopsVal.LessThan(decimal.NewFromInt(1000)) {
				iopsVal = decimal.NewFromInt(1000)
			}
			if allocatedStorageVal.LessThan(decimal.NewFromInt(100)) {
				allocatedStorageVal = decimal.NewFromInt(100)
			}
		} else if strings.ToLower(d.Get("storage_type").String()) == "standard" {
			volumeType = "Magnetic"
			storageName = "Storage (magnetic)"
		}
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
			Name:           fmt.Sprintf("Database instance (on-demand, %s, %s)", deploymentOption, instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
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
			Name:            storageName,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
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

	if strings.ToLower(volumeType) == "magnetic" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "I/O requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1000000),
			MonthlyQuantity: monthlyIORequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/RDS:StorageIOUsage/i"))},
				},
			},
		})
	}

	if strings.ToLower(volumeType) == "provisioned iops" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Provisioned IOPS",
			Unit:            "IOPS",
			UnitMultiplier:  decimal.NewFromInt(1),
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
