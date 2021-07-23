package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
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
	if strings.ToLower(engine) == "rabbitmq" {
		storageType = "ebs"
	}
	if d.Get("storage_type").Type != gjson.Null {
		storageType = strings.ToLower(d.Get("storage_type").String())
	}

	deploymentMode := "single_instance"
	if d.Get("deployment_mode").Type != gjson.Null {
		deploymentMode = strings.ToLower(d.Get("deployment_mode").String())
	}

	isMultiAZ := false
	if strings.ToLower(deploymentMode) == "active_standby_multi_az" || strings.ToLower(deploymentMode) == "cluster_multi_az" {
		isMultiAZ = true
	}

	var storageSizeGB *decimal.Decimal
	if u != nil && u.Get("storage_size_gb").Type != gjson.Null {
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
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Instances"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", instanceType))},
				{Key: "brokerEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", engine))},
				{Key: "deploymentOption", ValueRegex: strPtr(fmt.Sprintf("/%s/i", deploymentOption))},
			},
		},
	}
}

func storage(region, engine, storageType string, isMultiAZ bool, storageSizeGB *decimal.Decimal) *schema.CostComponent {
	instancesCount := decimalPtr(decimal.NewFromInt(1))
	usageType := ""

	if strings.ToLower(engine) == "rabbitmq" {
		usageType = "TimedStorage-RabbitMQ-ByteHrs"
		if isMultiAZ {
			instancesCount = decimalPtr(decimal.NewFromInt(3))
		}
	} else {
		// ActiveMQ
		if strings.ToLower(storageType) == "ebs" {
			usageType = "TimedStorage-EBS-ByteHrs"
		} else {
			// EFS
			usageType = "TimedStorage-ByteHrs"
		}
	}

	var summedStorageSizeGB *decimal.Decimal
	if storageSizeGB != nil {
		summedStorageSizeGB = decimalPtr(storageSizeGB.Mul(*instancesCount))
	}

	costComponent := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", engine, strings.ToUpper(storageType)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: summedStorageSizeGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
	}
	return costComponent
}
