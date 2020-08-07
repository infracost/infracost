package aws

import (
	"fmt"
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func rdsInstanceGbQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	sizeVal := resource.RawValues()["allocated_storage"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}

	return quantity
}

func rdsInstanceIopsQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	iopsVal := resource.RawValues()["iops"]
	if iopsVal != nil {
		quantity = decimal.NewFromFloat(iopsVal.(float64))
	}

	return quantity
}

func NewRdsInstance(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	deploymentOption := "Single-AZ"
	if rawValues["multi_az"] != nil && rawValues["multi_az"].(bool) {
		deploymentOption = "Multi-AZ"
	}

	instanceType := rawValues["instance_class"].(string)

	var databaseEngine string
	switch rawValues["engine"].(string) {
	case "postgresql":
		databaseEngine = "PostgreSQL"
	case "mysql":
		databaseEngine = "MySQL"
	case "mariadb":
		databaseEngine = "MariaDB"
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
	case "oracle-se", "oracle-se1", "oracle-se2", "oracle-ee":
		databaseEngine = "Oracle"
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		databaseEngine = "SQL Server"
	}

	var databaseEdition string
	switch rawValues["engine"].(string) {
	case "oracle-se", "sqlserver-se":
		databaseEdition = "Standard"
	case "oracle-se1":
		databaseEdition = "Standard One"
	case "oracle-se2":
		databaseEdition = "Standard 2"
	case "oracle-ee", "sqlserver-ee":
		databaseEdition = "Enterprise"
	case "sqlserver-ex":
		databaseEdition = "Express"
	case "sqlserver-web":
		databaseEdition = "Web"
	}

	volumeType := "General Purpose"
	if rawValues["storage_type"] == nil && rawValues["iops"] != nil {
		volumeType = "Provisioned IOPS"
	} else if rawValues["storage_type"] != nil {
		switch rawValues["storage_type"].(string) {
		case "standard":
			volumeType = "Magnetic"
		case "io1":
			volumeType = "Provisioned IOPS"
		}
	}

	var databaseEditionValue *string
	if databaseEdition != "" {
		databaseEditionValue = strPtr(databaseEdition)
	}
	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonRDS"),
		ProductFamily: strPtr("Database Instance"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "instanceType", Value: strPtr(instanceType)},
			{Key: "deploymentOption", Value: strPtr(deploymentOption)},
			{Key: "databaseEngine", Value: strPtr(databaseEngine)},
			{Key: "databaseEdition", Value: databaseEditionValue},
		},
	}
	hours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour", hoursProductFilter, nil)
	r.AddPriceComponent(hours)

	gbProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonRDS"),
		ProductFamily: strPtr("Database Storage"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "volumeType", Value: strPtr(volumeType)},
			{Key: "deploymentOption", Value: strPtr(deploymentOption)},
		},
	}
	gb := resource.NewBasePriceComponent("GB", r, "GB/month", "month", gbProductFilter, nil)
	gb.SetQuantityMultiplierFunc(rdsInstanceGbQuantity)
	r.AddPriceComponent(gb)

	if volumeType == "Provisioned IOPS" {
		iopsProductFilter := &resource.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Provisioned IOPS"),
			AttributeFilters: &[]resource.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
			},
		}
		iops := resource.NewBasePriceComponent("IOPS", r, "IOPS/month", "month", iopsProductFilter, nil)
		iops.SetQuantityMultiplierFunc(rdsInstanceIopsQuantity)
		r.AddPriceComponent(iops)
	}

	return r
}
