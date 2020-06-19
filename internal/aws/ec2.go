package aws

import (
	"plancosts/pkg/base"
)

var Ec2BlockDeviceGB = &base.PriceMapping{
	TimeUnit:       EbsVolumeGB.TimeUnit,
	DefaultFilters: EbsVolumeGB.DefaultFilters,
	CalculateCost:  EbsVolumeGB.CalculateCost,

	ValueMappings: []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	},
}

var Ec2BlockDeviceIOPS = &base.PriceMapping{
	TimeUnit:       EbsVolumeIOPS.TimeUnit,
	DefaultFilters: EbsVolumeIOPS.DefaultFilters,
	CalculateCost:  EbsVolumeIOPS.CalculateCost,

	ValueMappings: []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	},

	ShouldSkip: func(values map[string]interface{}) bool {
		return values["volume_type"] != "io1"
	},
}

var Ec2BlockDevice = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB":   Ec2BlockDeviceGB,
		"IOPS": Ec2BlockDeviceIOPS,
	},
}

var Ec2InstanceHours = &base.PriceMapping{
	TimeUnit: "hour",

	DefaultFilters: []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Compute Instance"},
		{Key: "operatingSystem", Value: "Linux"},
		{Key: "preInstalledSw", Value: "NA"},
		{Key: "capacitystatus", Value: "Used"},
		{Key: "tenancy", Value: "Shared"},
	},

	ValueMappings: []base.ValueMapping{
		{FromKey: "instance_type", ToKey: "instanceType"},
		{FromKey: "tenancy", ToKey: "tenancy"},
	},
}

var Ec2Instance = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"Instance hours": Ec2InstanceHours,
	},

	SubResourceMappings: map[string]*base.ResourceMapping{
		"root_block_device": Ec2BlockDevice,
		"ebs_block_device":  Ec2BlockDevice,
	},
}
