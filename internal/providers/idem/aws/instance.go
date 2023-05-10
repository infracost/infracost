package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "states.aws.ec2.instance.present",
		Notes: []string{
			"Costs associated with marketplace AMIs are not supported.",
			"For non-standard Linux AMIs such as Windows and RHEL, the operating system should be specified in usage file.",
			"EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.",
			"If a root volume is not specified then an 8Gi gp2 volume is assumed.",
		},
		CoreRFunc: NewInstance,
		ReferenceAttributes: []string{
			//"states.aws.ec2.instance.present:block_device_mappings.volume_id",
			"states.aws.ec2.volume.present:block_device_mappings.#.volume_id",
		},
	}
}

func NewInstance(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()

	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	var instanceType, ami, cpuCredits, tenancy string
	var ebsOptimized, monitoring bool
	ltEBSBlockDevices := map[string]*aws.EBSVolume{}

	launchTemplateRefs := d.References("states.aws.ec2.launch_template.present:launch_template.LaunchTemplateId")
	if len(launchTemplateRefs) == 0 {
		launchTemplateRefs = d.References("states.aws.ec2.launch_template.present:name")
	}

	if len(launchTemplateRefs) > 0 {
		ref := launchTemplateRefs[0]

		instanceType = ref.Get("instance_type").String()
		ami = ref.Get("image_id").String()
		ebsOptimized = ref.Get("ebs_optimized").Bool()
		monitoring = ref.Get("instance_monitoring.Enabled").Bool()
		cpuCredits = ref.Get("cpu_credits").String()
		tenancy = ref.Get("tenancy").String()

		for _, data := range ref.Get("block_device_mappings").Array() {
			deviceName := data.Get("device_name").String()
			ebsBlockDevice := &aws.EBSVolume{
				Region: region,
				Type:   data.Get("ebs.0.volume_type").String(),
				IOPS:   data.Get("ebs.0.iops").Int(),
			}

			if v := data.Get("ebs.0.volume_size"); v.Exists() {
				ebsBlockDevice.Size = intPtr(v.Int())
			}

			ltEBSBlockDevices[deviceName] = ebsBlockDevice
		}
	}

	instanceType = d.GetStringOrDefault("instance_type", instanceType)
	ami = d.GetStringOrDefault("ami", ami)
	ebsOptimized = d.GetBoolOrDefault("ebs_optimized", ebsOptimized)
	monitoring = d.GetBoolOrDefault("monitoring", monitoring)
	cpuCredits = d.GetStringOrDefault("cpu_credits", cpuCredits)
	tenancy = d.GetStringOrDefault("tenancy", tenancy)
	hasHost := len(d.References("host_id")) > 0

	a := &aws.Instance{
		Address:          d.Address,
		Region:           region,
		Tenancy:          tenancy,
		PurchaseOption:   purchaseOption,
		AMI:              ami,
		InstanceType:     instanceType,
		EBSOptimized:     ebsOptimized,
		EnableMonitoring: monitoring,
		CPUCredits:       cpuCredits,
		HasHost:          hasHost,
	}

	a.RootBlockDevice = &aws.EBSVolume{
		Address: "root_device_name",
		Region:  region,
		Type:    d.Get("root_block_device.0.volume_type").String(),
		IOPS:    d.Get("root_block_device.0.iops").Int(),
	}

	//if d.Get("root_block_device.0.volume_size").Type != gjson.Null {
	//	a.RootBlockDevice.Size = intPtr(d.Get("root_block_device.0.volume_size").Int())
	//}
	//
	//ebsBlockDeviceRef := d.References("ebs_block_device.#.volume_id")

	ebsBlockDeviceRef := d.References("states.aws.ec2.volume.present:block_device_mappings.#.volume_id")

	for i, data := range d.Get("block_device_mappings").Array() {
		deviceName := data.Get("device_name").String()

		ltDevice := ltEBSBlockDevices[deviceName]
		if ltDevice == nil {
			ltDevice = &aws.EBSVolume{}
		}

		// Check if there's a reference for this EBS volume
		// If there is then we shouldn't add as a subresource since
		// the cost will be added against the volume resource.
		//if len(ebsBlockDeviceRef) > i && ebsBlockDeviceRef[i].Get("resource_id").String() == data.Get("volume_id").String() {
		//	delete(ltEBSBlockDevices, deviceName)
		//
		//	continue
		//}

		// Instance can override individual Launch Template's values based on device
		// name.
		volumeType := ltDevice.Type
		if v := ebsBlockDeviceRef[i].Get("volume_type"); v.Exists() {
			volumeType = v.String()
		}

		volumeSize := ltDevice.Size
		if v := ebsBlockDeviceRef[i].Get("size"); v.Exists() {
			volumeSize = intPtr(v.Int())
		}

		iops := ltDevice.IOPS
		if v := ebsBlockDeviceRef[i].Get("iops"); v.Exists() {
			iops = v.Int()
		}

		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("ebs_block_device[%d]", i),
			Region:  region,
			Type:    volumeType,
			Size:    volumeSize,
			IOPS:    iops,
		}

		delete(ltEBSBlockDevices, deviceName)

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	// Add remaining EBS block devices from Launch Template.
	blockDevicesCount := len(a.EBSBlockDevices)
	for _, device := range ltEBSBlockDevices {
		device.Address = fmt.Sprintf("ebs_block_device[%d]", blockDevicesCount)
		a.EBSBlockDevices = append(a.EBSBlockDevices, device)
		blockDevicesCount++
	}

	return a
}
