package aws

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getECSServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ecs_service",
		RFunc:               NewECSService,
		ReferenceAttributes: []string{"task_definition"},
	}
}

func NewECSService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var taskDefinition *schema.ResourceData

	memoryGB := float64(0)
	vcpu := float64(0)
	inferenceAcceleratorDeviceType := ""

	taskDefinitionRefs := d.References("task_definition")
	if len(taskDefinitionRefs) > 0 {
		taskDefinition = taskDefinitionRefs[0]

		memoryGB = parseVCPUMemoryString(taskDefinition.Get("memory").String())
		vcpu = parseVCPUMemoryString(taskDefinition.Get("cpu").String())
		inferenceAcceleratorDeviceType = taskDefinition.Get("inference_accelerator.0.device_type").String()
	}

	r := &aws.ECSService{
		Address:                        d.Address,
		Region:                         d.Get("region").String(),
		LaunchType:                     d.Get("launch_type").String(),
		DesiredCount:                   d.Get("desired_count").Int(),
		MemoryGB:                       memoryGB,
		VCPU:                           vcpu,
		InferenceAcceleratorDeviceType: inferenceAcceleratorDeviceType,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
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
