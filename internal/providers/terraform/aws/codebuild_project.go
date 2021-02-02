package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCodebuildProjectRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_codebuild_project",
		RFunc: NewCodebuildProject,
	}
}

func NewCodebuildProject(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	environmentComputeType := d.Get("environment.0.compute_type").String()
	environmentType := d.Get("environment.0.type").String()

	var monthlyBuildMinutes int64
	if u != nil && u.Get("monthly_build_minutes").Exists() {
		monthlyBuildMinutes = u.Get("monthly_build_minutes").Int()
	}

	if environmentComputeType == "BUILD_GENERAL1_SMALL" {
		if monthlyBuildMinutes <= 100 {
			return &schema.Resource{
				NoPrice:   true,
				IsSkipped: true,
			}
		} else {
			monthlyBuildMinutes -= 100
		}
	}

	usageType := SetUsageType(environmentComputeType, environmentType)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            fmt.Sprintf("CodeBuild instance (%s)", usageType),
				Unit:            "minutes",
				UnitMultiplier:  1,
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(monthlyBuildMinutes)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("CodeBuild"),
					ProductFamily: strPtr("Compute"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", Value: strPtr(usageType)},
					},
				},
			},
		},
	}
}

func SetUsageType(environmentComputeType string, environmentType string) string {
	usageTypeTemplate := "USE1-Build-Min:"
	environmentType = SetValidEnvironmentType(environmentType)
	environmentComputeType = SetValidEnvironmentComputeType(environmentComputeType)

	return usageTypeTemplate + environmentType + environmentComputeType
}

func SetValidEnvironmentType(environmentType string) string {
	switch environmentType {
	case "LINUX_CONTAINER":
		return "Linux"
	case "LINUX_GPU_CONTAINER":
		return "LinuxGPU"
	case "ARM_CONTAINER":
		return "ARM"
	case "WINDOWS_SERVER_2019_CONTAINER":
		return "Windows"
	default:
		return ""
	}
}

func SetValidEnvironmentComputeType(computeType string) string {
	switch computeType {
	case "BUILD_GENERAL1_SMALL":
		return ":g1.small"
	case "BUILD_GENERAL1_MEDIUM":
		return ":g1.medium"
	case "BUILD_GENERAL1_LARGE":
		return ":g1.large"
	case "BUILD_GENERAL1_2XLARGE":
		return ":g1.2xlarge"
	default:
		return ""
	}
}
