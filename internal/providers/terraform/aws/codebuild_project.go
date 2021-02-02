package aws

import (
	"strings"

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

	computeType := d.Get("environment.0.compute_type").String()
	envType := d.Get("environment.0.type").String()

	var monthlyBuildMinutes int64
	if u != nil && u.Get("monthly_build_mins").Exists() {
		monthlyBuildMinutes = u.Get("monthly_build_mins").Int()
	}

	usageType := usageType(computeType, envType)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            amazonValidName(computeType, envType),
				Unit:            "minutes",
				UnitMultiplier:  1,
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(monthlyBuildMinutes)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("CodeBuild"),
					ProductFamily: strPtr("Compute"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(usageType)},
					},
				},
			},
		},
	}
}

func amazonValidName(computeType string, envType string) string {
	name := ""
	switch envType {
	case "WINDOWS_SERVER_2019_CONTAINER":
		name = "Windows ("
		name += strings.Replace(strings.ToLower(strings.SplitAfter(computeType, "BUILD_")[1]), "_", ".", 1) + ")"
	case "ARM_CONTAINER":
		name = "Linux (arm1.large)"
	case "LINUX_GPU_CONTAINER":
		name = "Linux (gpu1.large)"
	default:
		name = "Linux ("
		name += strings.Replace(strings.ToLower(strings.SplitAfter(computeType, "BUILD_")[1]), "_", ".", 1) + ")"
	}

	return name
}

func usageType(computeType string, envType string) string {
	envType = validEnvironmentType(envType)
	computeType = validEnvironmentComputeType(computeType)

	return "/" + envType + ":" + computeType + "/"
}

func validEnvironmentType(envType string) string {
	switch envType {
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

func validEnvironmentComputeType(computeType string) string {
	switch computeType {
	case "BUILD_GENERAL1_SMALL":
		return "g1.small"
	case "BUILD_GENERAL1_MEDIUM":
		return "g1.medium"
	case "BUILD_GENERAL1_LARGE":
		return "g1.large"
	case "BUILD_GENERAL1_2XLARGE":
		return "g1.2xlarge"
	default:
		return ""
	}
}
