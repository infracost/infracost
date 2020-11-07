package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func GetECSServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ecs_service",
		Notes:               []string{"Only supports Fargate on-demand."},
		RFunc:               NewECSService,
		ReferenceAttributes: []string{"task_definition"},
	}
}

func NewECSService(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	launchType := d.Get("launch_type").String()
	if launchType != "FARGATE" {
		log.Warnf("Skipping resource %s. Infracost currently only supports the FARGATE launch type for AWS ECS Services", d.Address)
		return nil
	}

	region := d.Get("region").String()
	desiredCount := int64(0)
	if d.Get("desired_count").Exists() {
		desiredCount = d.Get("desired_count").Int()
	}

	var taskDefinition *schema.ResourceData
	refs := d.References("task_definition")
	if len(refs) > 0 {
		taskDefinition = refs[0]
	}
	memory := decimal.Zero
	cpu := decimal.Zero
	if taskDefinition != nil {
		memory = convertResourceString(taskDefinition.Get("memory").String())
		cpu = convertResourceString(taskDefinition.Get("cpu").String())
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           "Per GB per hour",
			Unit:           "GB-hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredCount).Mul(memory)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-GB-Hours/")},
				},
			},
		},
		{
			Name:           "Per vCPU per hour",
			Unit:           "CPU-hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredCount).Mul(cpu)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-vCPU-Hours:perCPU/")},
				},
			},
		},
	}

	if taskDefinition != nil && taskDefinition.Get("inference_accelerator.0").Exists() {
		deviceType := taskDefinition.Get("inference_accelerator.0.device_type").String()
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Inference accelerator (%s)", deviceType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEI"),
				ProductFamily: strPtr("Elastic Inference"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", deviceType))},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func convertResourceString(rawValue string) decimal.Decimal {
	var quantity decimal.Decimal
	noSpaceString := strings.ReplaceAll(rawValue, " ", "")
	reg := regexp.MustCompile(`(?i)vcpu|gb`)
	if reg.MatchString(noSpaceString) {
		quantity, _ = decimal.NewFromString(reg.ReplaceAllString(noSpaceString, ""))
	} else {
		quantity, _ = decimal.NewFromString(noSpaceString)
		quantity = quantity.Div(decimal.NewFromInt(1024))
	}
	return quantity
}
