package aws

import (
	"fmt"

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
	deploymentMode := d.Get("deployment_mode").String()
	isActiveStandby := false
	if deploymentMode == "ACTIVE_STANDBY_MULTI_AZ" {
		isActiveStandby = true
	}

	nodesCount := decimalPtr(decimal.NewFromInt(1))
	if u != nil && u.Get("nodes_count").Exists() {
		nodesCount = decimalPtr(decimal.NewFromInt(u.Get("nodes_count").Int()))
	}
	var storageSizeGB *decimal.Decimal
	if u != nil && u.Get("storage_size_gb").Exists() {
		storageSizeGB = decimalPtr(decimal.NewFromInt(u.Get("storage_size_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			instance(region, engine, instanceType, isActiveStandby, nodesCount),
			storage(region, engine, isActiveStandby, storageSizeGB, nodesCount),
		},
	}
}

func instance(region, engine, instanceType string, isActiveStandby bool, nodesCount *decimal.Decimal) *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if isActiveStandby {
		deploymentOption = "Multi-AZ"
	}
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance (%s, %s)", engine, instanceType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: nodesCount,
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

func storage(region, engine string, isActiveStandby bool, storageSizeGB, nodesCount *decimal.Decimal) *schema.CostComponent {
	if storageSizeGB != nil {
		freeSizeGB := decimal.NewFromInt(5)
		storageSizeGB = decimalPtr(storageSizeGB.Sub(freeSizeGB))
		if storageSizeGB.LessThanOrEqual(decimal.Zero) {
			storageSizeGB = decimalPtr(decimal.Zero)
		}
	}
	var summedStorageSizeGB *decimal.Decimal
	if storageSizeGB != nil {
		summedStorageSizeGB = decimalPtr(storageSizeGB.Mul(*nodesCount))
	}

	deploymentOption := "Single-AZ"
	if isActiveStandby {
		deploymentOption = "Multi-AZ"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", engine, deploymentOption),
		Unit:            "GB-months",
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
		PriceFilter: &schema.PriceFilter{
			DescriptionRegex: strPtr(fmt.Sprintf("/%s/", engine)),
		},
	}
}
