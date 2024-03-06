package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type ECSService struct {
	Address                        string
	LaunchType                     string
	Region                         string
	DesiredCount                   int64
	MemoryGB                       float64
	VCPU                           float64
	InferenceAcceleratorDeviceType string
}

func (r *ECSService) CoreType() string {
	return "ECSService"
}

func (r *ECSService) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ECSService) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ECSService) BuildResource() *schema.Resource {
	if r.LaunchType != "FARGATE" {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			NoPrice:     true,
			UsageSchema: r.UsageSchema(),
		}
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           "Per GB per hour",
			Unit:           "GB",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromFloat(r.MemoryGB * float64(r.DesiredCount))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-GB-Hours/")},
				},
			},
		},
		{
			Name:           "Per vCPU per hour",
			Unit:           "CPU",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(decimal.NewFromFloat(r.VCPU * float64(r.DesiredCount))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-vCPU-Hours:perCPU/")},
				},
			},
		},
	}

	if r.InferenceAcceleratorDeviceType != "" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Inference accelerator (%s)", r.InferenceAcceleratorDeviceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(r.DesiredCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonEI"),
				ProductFamily: strPtr("Elastic Inference"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", r.InferenceAcceleratorDeviceType))},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
