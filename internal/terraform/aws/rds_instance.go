package aws

import (
	"fmt"
	"infracost/pkg/base"

	"github.com/shopspring/decimal"
)

func rdsInstanceGbQuantity(resource base.Resource) decimal.Decimal {
	quantity := decimal.Zero

	sizeVal := resource.RawValues()["allocated_storage"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}

	return quantity
}

func rdsInstanceIopsQuantity(resource base.Resource) decimal.Decimal {
	quantity := decimal.Zero

	iopsVal := resource.RawValues()["iops"]
	if iopsVal != nil {
		quantity = decimal.NewFromFloat(iopsVal.(float64))
	}

	return quantity
}

func NewRdsInstance(address string, region string, rawValues map[string]interface{}) base.Resource {
	r := base.NewBaseResource(address, rawValues, true)

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
	if rawValues["storage_ty[e"] != nil {
		switch rawValues["storage_type"].(string) {
		case "standard":
			volumeType = "Magnetic"
		case "io1":
			volumeType = "Provisioned IOPS"
		}
	}

	hours := base.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonRDS"},
		{Key: "productFamily", Value: "Database Instance"},
		{Key: "deploymentOption", Value: deploymentOption},
		{Key: "instanceType", Value: instanceType},
		{Key: "databaseEngine", Value: databaseEngine},
	})
	if databaseEdition != "" {
		hours.AddFilters([]base.Filter{
			{Key: "databaseEdition", Value: databaseEdition},
		})
	}
	r.AddPriceComponent(hours)

	gb := base.NewBasePriceComponent("GB", r, "GB/month", "month")
	gb.AddFilters(regionFilters(region))
	gb.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonRDS"},
		{Key: "productFamily", Value: "Database Storage"},
		{Key: "deploymentOption", Value: deploymentOption},
		{Key: "volumeType", Value: volumeType},
	})
	gb.SetQuantityMultiplierFunc(rdsInstanceGbQuantity)
	r.AddPriceComponent(gb)

	if volumeType == "io1" {
		iops := base.NewBasePriceComponent("IOPS", r, "IOPS/month", "month")
		iops.AddFilters(regionFilters(region))
		iops.AddFilters([]base.Filter{
			{Key: "servicecode", Value: "AmazonRDS"},
			{Key: "productFamily", Value: "Provisioned IOPS"},
			{Key: "deploymentOption", Value: deploymentOption},
		})
		iops.SetQuantityMultiplierFunc(rdsInstanceIopsQuantity)
		r.AddPriceComponent(iops)
	}

	return r
}
