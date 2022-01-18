package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type CodebuildProject struct {
	Address                 *string
	Region                  *string
	Environment0ComputeType *string
	Environment0Type        *string
	MonthlyBuildMins        *int64 `infracost_usage:"monthly_build_mins"`
}

var CodebuildProjectUsageSchema = []*schema.UsageItem{{Key: "monthly_build_mins", ValueType: schema.Int64, DefaultValue: 0}}

func (r *CodebuildProject) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CodebuildProject) BuildResource() *schema.Resource {
	region := *r.Region

	computeType := *r.Environment0ComputeType
	envType := *r.Environment0Type

	var monthlyBuildMinutes *decimal.Decimal
	if r.MonthlyBuildMins != nil {
		monthlyBuildMinutes = decimalPtr(decimal.NewFromInt(*r.MonthlyBuildMins))
	}

	usageType := codeBuildUsageType(computeType, envType)

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            codeBuildNameLabel(computeType, envType),
				Unit:            "minutes",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyBuildMinutes,
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
		}, UsageSchema: CodebuildProjectUsageSchema,
	}
}

func codeBuildNameLabel(computeType string, envType string) string {
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

func codeBuildUsageType(computeType string, envType string) string {
	envType = codeBuildEnvType(envType)
	computeType = codeBuildComputeType(computeType)

	return "/" + envType + ":" + computeType + "/"
}

func codeBuildEnvType(envType string) string {
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

func codeBuildComputeType(computeType string) string {
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
