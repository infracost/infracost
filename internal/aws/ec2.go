package aws

import "plancosts/pkg/base"

var BlockDeviceGB = &base.PriceMapping{
	TimeUnit: "hour",

	ValueMappings: []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	},
}

var BlockDeviceIOPS = &base.PriceMapping{
	TimeUnit: "hour",

	ValueMappings: []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	},

	ShouldSkip: func(values map[string]interface{}) bool {
		return values["volume_type"] != "io1"
	},
}

var BlockDevice = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB":   BlockDeviceGB,
		"IOPS": BlockDeviceIOPS,
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
		"root_block_device": BlockDevice,
		"ebs_block_device":  BlockDevice,
	},
}
