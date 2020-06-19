package aws

import (
	"fmt"
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

var Ec2BlockDeviceMappingGB = &base.PriceMapping{
	TimeUnit:       Ec2BlockDeviceGB.TimeUnit,
	DefaultFilters: Ec2BlockDeviceGB.DefaultFilters,

	OverrideFilters: func(resource base.Resource) []base.Filter {
		volumeType := "gp2"
		volumeTypeVal := resource.RawValues()["ebs"].([]interface{})[0].(map[string]interface{})["volume_type"]
		if volumeTypeVal != nil {
			volumeType = volumeTypeVal.(string)
		}
		return []base.Filter{
			{Key: "volumeApiName", Value: volumeType},
		}
	},

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		size := decimal.NewFromInt(int64(0))
		sizeVal := resource.RawValues()["ebs"].([]interface{})[0].(map[string]interface{})["volume_size"]
		if sizeVal != nil {
			size = decimal.NewFromFloat(sizeVal.(float64))
		}
		return price.Mul(size)
	},
}

var Ec2BlockDeviceMappingIOPS = &base.PriceMapping{
	TimeUnit:       Ec2BlockDeviceIOPS.TimeUnit,
	DefaultFilters: Ec2BlockDeviceIOPS.DefaultFilters,

	OverrideFilters: func(resource base.Resource) []base.Filter {
		volumeType := "gp2"
		volumeTypeVal := resource.RawValues()["ebs"].([]interface{})[0].(map[string]interface{})["volume_type"]
		if volumeTypeVal != nil {
			volumeType = volumeTypeVal.(string)
		}
		return []base.Filter{
			{Key: "volumeApiName", Value: volumeType},
		}
	},

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		iops := decimal.NewFromInt(int64(0))
		iopsVal := resource.RawValues()["ebs"].([]interface{})[0].(map[string]interface{})["iops"]
		if iopsVal != nil {
			iops = decimal.NewFromFloat(iopsVal.(float64))
		}
		return price.Mul(iops)
	},

	ShouldSkip: func(values map[string]interface{}) bool {
		volumeType := values["ebs"].([]interface{})[0].(map[string]interface{})["volume_type"]
		return volumeType == nil || volumeType.(string) != "io1"
	},
}

var Ec2BlockDeviceMapping = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB":   Ec2BlockDeviceMappingGB,
		"IOPS": Ec2BlockDeviceMappingIOPS,
	},
}

var AutoscalingGroupLaunchConfigurationInstanceHours = &base.PriceMapping{
	TimeUnit:       Ec2InstanceHours.TimeUnit,
	DefaultFilters: Ec2InstanceHours.DefaultFilters,
	CalculateCost:  Ec2InstanceHours.CalculateCost,

	OverrideFilters: func(resource base.Resource) []base.Filter {
		filters := []base.Filter{
			{Key: "instanceType", Value: resource.References()["launch_configuration"].RawValues()["instance_type"].(string)},
		}
		placementTenancy := resource.References()["launch_configuration"].RawValues()["placement_tenancy"]
		if placementTenancy != nil {
			filters = append(filters, base.Filter{
				Key:   "tenancy",
				Value: placementTenancy.(string),
			})
		}
		return filters
	},
}

var AutoscalingGroupLaunchConfiguration = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"Instance hours": AutoscalingGroupLaunchConfigurationInstanceHours,
	},

	SubResourceMappings: map[string]*base.ResourceMapping{
		"block_device_mappings": Ec2BlockDeviceMapping,
	},

	OverrideSubResourceRawValues: func(resource base.Resource) map[string][]interface{} {
		rawValues := make(map[string][]interface{})
		blockDeviceMappings := resource.References()["launch_configuration"].RawValues()["block_device_mappings"]
		if blockDeviceMappings != nil {
			rawValues["block_device_mappings"] = blockDeviceMappings.([]interface{})
		}
		return rawValues
	},
}

var AutoscalingGroupLaunchTemplateInstanceHours = &base.PriceMapping{
	TimeUnit:       Ec2InstanceHours.TimeUnit,
	DefaultFilters: Ec2InstanceHours.DefaultFilters,
	CalculateCost:  Ec2InstanceHours.CalculateCost,

	OverrideFilters: func(resource base.Resource) []base.Filter {
		return []base.Filter{
			{Key: "instanceType", Value: resource.References()["launch_template"].RawValues()["instance_type"].(string)},
		}
	},
}

var AutoscalingGroupLaunchTemplate = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"Instance hours": AutoscalingGroupLaunchTemplateInstanceHours,
	},

	SubResourceMappings: map[string]*base.ResourceMapping{
		"block_device_mappings": Ec2BlockDeviceMapping,
	},

	OverrideSubResourceRawValues: func(resource base.Resource) map[string][]interface{} {
		rawValues := make(map[string][]interface{})
		blockDeviceMappings := resource.References()["launch_template"].RawValues()["block_device_mappings"]
		if blockDeviceMappings != nil {
			rawValues["block_device_mappings"] = blockDeviceMappings.([]interface{})
		}
		return rawValues
	},

	AdjustCost: func(resource base.Resource, cost decimal.Decimal) decimal.Decimal {
		count := decimal.NewFromInt(int64(1))
		if countStr, ok := resource.RawValues()["desired_capacity"]; ok {
			count, _ = decimal.NewFromString(fmt.Sprintf("%v", countStr))
		}
		return cost.Mul(count)
	},
}
