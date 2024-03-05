package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// LambdaProvisionedConcurrencyConfig initializes a requested number of execution environments so that
// they are prepared to respond immediately to your functions invocations. Configuring provisioned
// concurrency incurs charges to your AWS Account.
//
// Resource information: https://docs.aws.amazon.com/lambda/latest/dg/lambda-concurrency.html
// Pricing information: https://aws.amazon.com/lambda/pricing/#Provisioned_Concurrency_Pricing
type LambdaProvisionedConcurrencyConfig struct {
	Address                         string
	Region                          string
	Name                            string
	ProvisionedConcurrentExecutions int64

	MonthlyDurationHours *int64  `infracost_usage:"monthly_duration_hrs"`
	MonthlyRequests      *int64  `infracost_usage:"monthly_requests"`
	RequestDurationMS    *int64  `infracost_usage:"request_duration_ms"`
	Architecture         *string `infracost_usage:"architecture"`
	MemoryMB             *int64  `infracost_usage:"memory_mb"`
}

func (r *LambdaProvisionedConcurrencyConfig) CoreType() string {
	return "LambdaProvisionedConcurrencyConfig"
}

func (r *LambdaProvisionedConcurrencyConfig) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "memory_mb", ValueType: schema.Int64, DefaultValue: 512},
		{Key: "architecture", ValueType: schema.String, DefaultValue: "x86_64"},
		{Key: "monthly_duration_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "request_duration_ms", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *LambdaProvisionedConcurrencyConfig) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LambdaProvisionedConcurrencyConfig) BuildResource() *schema.Resource {
	monthlyDurationHours := decimal.NewFromInt(0)
	memorySize := decimal.NewFromInt(512)
	monthlyRequests := decimal.NewFromInt(0)

	concurrentExecutions := decimal.NewFromInt(r.ProvisionedConcurrentExecutions)

	if r.MonthlyRequests != nil {
		monthlyRequests = decimal.NewFromInt(*r.MonthlyRequests)
	}

	if r.MonthlyDurationHours != nil {
		monthlyDurationHours = decimal.NewFromInt(*r.MonthlyDurationHours)
	}

	averageRequestDuration := decimal.NewFromInt(1)
	if r.RequestDurationMS != nil {
		averageRequestDuration = decimal.NewFromInt(*r.RequestDurationMS)
	}

	if r.MemoryMB != nil {
		memorySize = decimal.NewFromInt(*r.MemoryMB)
	}

	totalSeconds := monthlyDurationHours.Mul(decimal.NewFromInt(3600))
	totalConcurrencyConfigured := concurrentExecutions.Mul(memorySize.Div(decimal.NewFromInt(1024)))
	totalConcurrency := totalConcurrencyConfigured.Mul(totalSeconds)

	concurrencyType := "AWS-Lambda-Provisioned-Concurrency"
	durationType := "AWS-Lambda-Duration-Provisioned"
	requestType := "AWS-Lambda-Requests"

	if strVal(r.Architecture) == "arm64" {
		concurrencyType = "AWS-Lambda-Provisioned-Concurrency-ARM"
		durationType = "AWS-Lambda-Duration-Provisioned-ARM"
		requestType = "AWS-Lambda-Requests-ARM"
	}

	provisionDuration := calculateGBSeconds(memorySize, averageRequestDuration, monthlyRequests)

	costComponents := []*schema.CostComponent{
		{
			Name:            "Requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1000000),
			MonthlyQuantity: decimalPtr(monthlyRequests),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AWSLambda"),
				ProductFamily: strPtr("Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr(requestType)},
					{Key: "usagetype", ValueRegex: strPtr("/Request/")},
				},
			},
			UsageBased: true,
		},
		{
			Name:            "Provisioned Concurrency",
			Unit:            "GB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &totalConcurrency,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AWSLambda"),
				ProductFamily: strPtr("Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr(concurrencyType)},
				},
			},
			UsageBased: true,
		},
		{
			Name:            "Duration",
			Unit:            "GB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &provisionDuration,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AWSLambda"),
				ProductFamily: strPtr("Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr(durationType)},
				},
			},
			UsageBased: true,
		},
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
