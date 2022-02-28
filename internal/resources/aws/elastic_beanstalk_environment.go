package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

// ElasticBeanstalkEnvironment struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/elasticbeanstalk/
// Pricing information: https://aws.amazon.com/elasticbeanstalk/pricing/
type ElasticBeanstalkEnvironment struct {
	Address               string
	Region                string
	InstanceType          string
	LoadBalancerType      string
	Nodes                 int64
	PurchaseOption        string
	StreamLogs            bool
	MonthlyDataIngestedGB *float64 `infracost_usage:"monthly_data_ingested_gb"`
	StorageGB             *float64 `infracost_usage:"storage_gb"`
	MonthlyDataScannedGB  *float64 `infracost_usage:"monthly_data_scanned_gb"`
}

// ElasticBeanstalkEnvironmentUsageSchema defines a list which represents the usage schema of ElasticBeanstalkEnvironment.
var ElasticBeanstalkEnvironmentUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_ingested_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_scanned_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the ElasticBeanstalkEnvironment.
// It uses the `infracost_usage` struct tags to populate data into the ElasticBeanstalkEnvironment.
func (r *ElasticBeanstalkEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ElasticBeanstalkEnvironment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ElasticBeanstalkEnvironment) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, r.instanceCostComponent())

	// Load balancer definition type
	if strings.ToLower(r.LoadBalancerType) == "application" {
		costComponents = append(costComponents, r.applicationLBCostComponents())
	} else if strings.ToLower(r.LoadBalancerType) == "network" {
		costComponents = append(costComponents, r.networkLBCostComponents())
	} else {
		costComponents = append(costComponents, r.classicLBCostComponents())
	}

	// Log streaming to cloudwatch
	if r.StreamLogs {
		costComponents = append(
			costComponents,
			r.cloudwatchDataIngestionCostComponents(),
			r.cloudwatchArchivalStorageCostComponents(),
			r.cloudwatchInsightsCostComponents(),
		)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ElasticBeanstalkEnvironmentUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *ElasticBeanstalkEnvironment) instanceCostComponent() *schema.CostComponent {

	osLabel := "Linux/UNIX"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, r.PurchaseOption, r.InstanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.Nodes)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(r.InstanceType)},
				{Key: "tenancy", Value: strPtr("Dedicated")},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "capacitystatus", Value: strPtr("Used")},
				{Key: "operatingSystem", Value: strPtr("Linux")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(r.PurchaseOption),
		},
	}
}

func (r *ElasticBeanstalkEnvironment) applicationLBCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:           "Application load balancer",
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr("Load Balancer-Application"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
			},
		},
	}
}

func (r *ElasticBeanstalkEnvironment) networkLBCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:           "Network load balancer",
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr("Load Balancer-Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
			},
		},
	}
}

func (r *ElasticBeanstalkEnvironment) classicLBCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:           "Classic load balancer",
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr("Load Balancer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
			},
		},
	}
}

func (r *ElasticBeanstalkEnvironment) cloudwatchDataIngestionCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:            "Cloudwatch Data ingested",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataIngestedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Data Payload"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/-DataProcessing-Bytes/")},
			},
		},
	}
}

func (r *ElasticBeanstalkEnvironment) cloudwatchArchivalStorageCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:            "Cloudwatch Archival Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.StorageGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-ByteHrs/")},
			},
		},
	}
}

func (r *ElasticBeanstalkEnvironment) cloudwatchInsightsCostComponents() *schema.CostComponent {

	return &schema.CostComponent{
		Name:            "Cloudwatch Insights queries data scanned",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataScannedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Data Payload"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/-DataScanned-Bytes/")},
			},
		},
	}
}
