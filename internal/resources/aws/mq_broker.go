package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type MQBroker struct {
	Address          string
	StorageType      string
	DeploymentMode   string
	Region           string
	EngineType       string
	HostInstanceType string
	StorageSizeGb    *float64 `infracost_usage:"storage_size_gb"`
}

func (r *MQBroker) CoreType() string {
	return "MQBroker"
}

func (r *MQBroker) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "storage_size_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *MQBroker) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MQBroker) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.instanceUsageCostComponent(),
			r.storageCostComponent(),
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *MQBroker) isMultiAZ() bool {
	if strings.ToLower(r.DeploymentMode) == "active_standby_multi_az" || strings.ToLower(r.DeploymentMode) == "cluster_multi_az" {
		return true
	}

	return false
}

func (r *MQBroker) instanceUsageCostComponent() *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if r.isMultiAZ() {
		deploymentOption = "Multi-AZ"
	}

	deploymentMode := strings.ToLower(r.DeploymentMode)
	if deploymentMode == "" {
		deploymentMode = "single_instance"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", r.EngineType, r.HostInstanceType, strings.ToLower(deploymentMode)),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Instances"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr(r.HostInstanceType)},
				{Key: "brokerEngine", ValueRegex: regexPtr(r.EngineType)},
				{Key: "deploymentOption", ValueRegex: regexPtr(deploymentOption)},
				{Key: "operation", ValueRegex: regexPtr("CreateBroker")},
			},
		},
	}
}

func (r *MQBroker) storageCostComponent() *schema.CostComponent {
	instanceCount := int64(1)
	if strings.ToLower(r.EngineType) == "rabbitmq" && r.isMultiAZ() {
		instanceCount = int64(3)
	}

	storageType := strings.ToLower(r.StorageType)
	if storageType == "" {
		if strings.ToLower(r.EngineType) == "rabbitmq" {
			storageType = "ebs"
		} else {
			storageType = "efs"
		}
	}

	usageType := "TimedStorage-ByteHrs"
	if strings.ToLower(r.EngineType) == "rabbitmq" {
		usageType = "TimedStorage-RabbitMQ-ByteHrs"
	} else if strings.ToLower(storageType) == "ebs" {
		usageType = "TimedStorage-EBS-ByteHrs"
	}

	var storageSizeGB *decimal.Decimal
	if r.StorageSizeGb != nil {
		storageSizeGB = decimalPtr(decimal.NewFromFloat(*r.StorageSizeGb).Mul(decimal.NewFromInt(instanceCount)))
	}

	costComponent := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", r.EngineType, strings.ToUpper(storageType)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageSizeGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonMQ"),
			ProductFamily: strPtr("Broker Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		UsageBased: true,
	}
	return costComponent
}
