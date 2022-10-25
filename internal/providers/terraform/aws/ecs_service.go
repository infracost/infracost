package aws

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getECSServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ecs_service",
		RFunc:               NewECSService,
		ReferenceAttributes: []string{"cluster", "task_definition"},
	}
}

func NewECSService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	memoryGB := float64(0)
	vcpu := float64(0)
	inferenceAcceleratorDeviceType := ""

	var taskDefinition *schema.ResourceData
	// Since we are matching on 'family' as well as 'arn', check the resource type
	// just in case the reference matches other resources as well. We should probably specify the
	// expected resource types when we are building the references, but we don;t just now so
	// this check should be sufficient.
	for _, ref := range d.References("task_definition") {
		if ref.Type == "aws_ecs_task_definition" {
			taskDefinition = ref
			break
		}
	}

	if taskDefinition != nil {
		memoryGB = parseVCPUMemoryString(taskDefinition.Get("memory").String())
		vcpu = parseVCPUMemoryString(taskDefinition.Get("cpu").String())
		inferenceAcceleratorDeviceType = taskDefinition.Get("inference_accelerator.0.device_type").String()
	}

	r := &aws.ECSService{
		Address:                        d.Address,
		Region:                         d.Get("region").String(),
		LaunchType:                     calcLaunchType(d),
		DesiredCount:                   d.Get("desired_count").Int(),
		MemoryGB:                       memoryGB,
		VCPU:                           vcpu,
		InferenceAcceleratorDeviceType: inferenceAcceleratorDeviceType,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}

// calcLaunchType determines the launch type for the resource using the following precedence:
//   1. aws_ecs_service.launch_type
//   2. aws_ecs_service.capacity_provider_strategy
//   3. aws_ecs_service.aws_ecs_cluster.default_capacity_provider_strategy
//   4. aws_ecs_service.aws_ecs_cluster.aws_ecs_cluster_capacity_providers
func calcLaunchType(d *schema.ResourceData) string {
	// Use the launch_type if it is set
	launchType := d.Get("launch_type").String()
	if launchType != "" {
		return launchType
	}

	// Check for an active direct capacity provider
	launchType = getCapacityProviderLaunchType(d.Get("capacity_provider_strategy").Array())
	if launchType != "" {
		return launchType
	}

	clusterRefs := d.References("cluster")
	if len(clusterRefs) > 0 {
		cluster := clusterRefs[0]

		// check the cluster for a default capacity provider

		if defaultStrategies := cluster.Get("default_capacity_provider_strategy").Array(); len(defaultStrategies) > 0 {
			launchType = getCapacityProviderLaunchType(defaultStrategies)
			if launchType == "FARGATE" {
				return launchType
			}
		} else {
			// since there are no default strategies, look for a directly set fargate capacity provider
			for _, capProvider := range cluster.Get("capacity_providers").Array() {
				if capProvider.String() == "FARGATE" {
					return "FARGATE"
				}
			}
		}

		// check for aws_ecs_cluster_capacity_providers
		for _, capProvider := range cluster.References("aws_ecs_cluster_capacity_providers.cluster_name") {
			defaultStrategies := capProvider.Get("default_capacity_provider_strategy").Array()
			if len(defaultStrategies) > 0 {
				lt := getCapacityProviderLaunchType(defaultStrategies)
				if lt == "FARGATE" {
					return lt
				}
				if lt != "" {
					launchType = lt
				}
			} else {
				// since there are no default strategies, look for a directly set fargate capacity provider
				for _, capProvider := range capProvider.Get("capacity_providers").Array() {
					if capProvider.String() == "FARGATE" {
						return "FARGATE"
					}
				}
			}

		}
	}

	return launchType
}

func getCapacityProviderLaunchType(capacityProviderStrategies []gjson.Result) string {
	launchType := ""
	for _, data := range capacityProviderStrategies {
		provider := strings.ToUpper(data.Get("capacity_provider").String())
		base := data.Get("base").Int()
		weight := data.Get("weight").Int()
		if base > 0 || weight > 0 {
			if strings.HasPrefix(provider, "FARGATE") {
				// We have at least one fargate provider, use that as the launch type
				return "FARGATE"
			} else {
				launchType = "EC2"
			}
		}
	}
	return launchType
}

func parseVCPUMemoryString(rawValue string) float64 {
	var quantity float64

	noSpaceString := strings.ReplaceAll(rawValue, " ", "")

	reg := regexp.MustCompile(`(?i)vcpu|gb`)
	if reg.MatchString(noSpaceString) {
		quantity, _ = strconv.ParseFloat(reg.ReplaceAllString(noSpaceString, ""), 64)
	} else {
		quantity, _ = strconv.ParseFloat(noSpaceString, 64)
		quantity /= 1024.0
	}

	return quantity
}
