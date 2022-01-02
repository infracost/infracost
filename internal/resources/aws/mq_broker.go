package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type MqBroker struct {
	Address          *string
	StorageType      *string
	DeploymentMode   *string
	Region           *string
	EngineType       *string
	HostInstanceType *string
	StorageSizeGb    *float64 `infracost_usage:"storage_size_gb"`
}

var MqBrokerUsageSchema = []*schema.UsageItem{{Key: "storage_size_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *MqBroker) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MqBroker) BuildResource() *schema.Resource {
	region := *r.Region
	engine := *r.EngineType
	instanceType := *r.HostInstanceType

	storageType := "efs"
	if strings.ToLower(engine) == "rabbitmq" {
		storageType = "ebs"
	}
	if r.StorageType != nil {
		storageType = strings.ToLower(*r.StorageType)
	}

	deploymentMode := "single_instance"
	if r.DeploymentMode != nil {
		deploymentMode = strings.ToLower(*r.DeploymentMode)
	}

	isMultiAZ := false
	if strings.ToLower(deploymentMode) == "active_standby_multi_az" || strings.ToLower(deploymentMode) == "cluster_multi_az" {
		isMultiAZ = true
	}

	var storageSizeGB *decimal.Decimal
	if r.StorageSizeGb != nil {
		storageSizeGB = decimalPtr(decimal.NewFromFloat(*r.StorageSizeGb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			instance(region, engine, instanceType, deploymentMode, isMultiAZ),
			storage(region, engine, storageType, isMultiAZ, storageSizeGB),
		}, UsageSchema: MqBrokerUsageSchema,
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

		if strings.ToLower(storageType) == "ebs" {
			usageType = "TimedStorage-EBS-ByteHrs"
		} else {

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
