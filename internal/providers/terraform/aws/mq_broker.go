package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetMQBrokerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_mq_broker",
		RFunc: NewMQBroker,
	}
}

func NewMQBroker(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	engine := d.Get("engine_type").String()
	instanceType := d.Get("host_instance_type").String()
	storageType := "efs"
	if d.Get("storage_type").Exists() {
		storageType = d.Get("storage_type").String()
	}
	deploymentMode := d.Get("deployment_mode").String()
	isMultiAZ := false
	if deploymentMode == "ACTIVE_STANDBY_MULTI_AZ" || deploymentMode == "CLUSTER_MULTI_AZ" {
		isMultiAZ = true
	}

	var storageSizeGB *decimal.Decimal
	if u != nil && u.Get("storage_size_gb").Exists() {
		storageSizeGB = decimalPtr(decimal.NewFromInt(u.Get("storage_size_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			instance(region, engine, instanceType, deploymentMode, isMultiAZ),
			storage(region, engine, storageType, isMultiAZ, storageSizeGB),
		},
	}
}

func instance(region, engine, instanceType, deploymentMode string, isMultiAZ bool) *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if isMultiAZ {
		deploymentOption = "Multi-AZ"
	}
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", engine, instanceType, strings.ToLower(deploymentMode)),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Instances"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", instanceType))},
				{Key: "brokerEngine", Value: strPtr(engine)},
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
			},
		},
	}
}

func storage(region, engine, storageType string, isMultiAZ bool, storageSizeGB *decimal.Decimal) *schema.CostComponent {

	instancesCount := decimalPtr(decimal.NewFromInt(1))
	if engine == "RabbitMQ" {
		storageType = "ebs"
		instancesCount = decimalPtr(decimal.NewFromInt(3))
	} else if isMultiAZ {
		// ActiveMQ
		instancesCount = decimalPtr(decimal.NewFromInt(2))
	}
	deploymentOption := "Single-AZ"
	if engine != "RabbitMQ" && storageType == "efs" {
		deploymentOption = "Multi-AZ"
	}

	var summedStorageSizeGB *decimal.Decimal
	if storageSizeGB != nil {
		summedStorageSizeGB = decimalPtr(storageSizeGB.Mul(*instancesCount))
	}

	costComponent := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", engine, strings.ToUpper(storageType)),
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: summedStorageSizeGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
			},
		},
	}
	if engine == "RabbitMQ" {
		costComponent.ProductFilter.AttributeFilters = append(costComponent.ProductFilter.AttributeFilters, &schema.AttributeFilter{
			Key:        "usagetype",
			ValueRegex: strPtr("/RabbitMQ/"),
		})
	}
	return costComponent
}
