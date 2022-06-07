package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func getInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		//Name: "aws:ec2/instance:Instance",
		Name: "aws_ec2_instance",
		Notes: []string{
			"Costs associated with marketplace AMIs are not supported.",
			"For non-standard Linux AMIs such as Windows and RHEL, the operating system should be specified in usage file.",
			"EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.",
			"If a root volume is not specified then an 8Gi gp2 volume is assumed.",
		},
		RFunc: NewInstance,
		ReferenceAttributes: []string{
			"ebsBlockDevices.#.volumeId",
			"launchTemplate.#.id",
			"launchTemplate.#.name",
		},
	}
}

func NewInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var region = d.Get("config.aws:region")
	log.Debugf("resources %s", d)
	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	var instanceType, ami, cpuCredits, tenancy string
	var ebsOptimized, monitoring bool
	ltEBSBlockDevices := map[string]*aws.EBSVolume{}

	launchTemplateRefs := d.References("launchTemplate.#.id")
	if len(launchTemplateRefs) == 0 {
		launchTemplateRefs = d.References("launchTemplate.#.name")
	}

	if len(launchTemplateRefs) > 0 {
		ref := launchTemplateRefs[0]

		instanceType = ref.Get("instanceType").String()
		ami = ref.Get("image_id").String()
		ebsOptimized = ref.Get("ebs_optimized").Bool()
		monitoring = ref.Get("monitoring.0.enabled").Bool()
		cpuCredits = ref.Get("credit_specification.0.cpu_credits").String()
		tenancy = ref.Get("placement.0.tenancy").String()

		for _, data := range ref.Get("ebsBlockDevices").Array() {
			deviceName := data.Get("deviceName").String()
			ebsBlockDevice := &aws.EBSVolume{
				Region: region.String(),
				Type:   data.Get("ebs.0.volume_type").String(),
				IOPS:   data.Get("ebs.0.iops").Int(),
			}

			if v := data.Get("ebs.0.volume_size"); v.Exists() {
				ebsBlockDevice.Size = intPtr(v.Int())
			}

			ltEBSBlockDevices[deviceName] = ebsBlockDevice
		}
	}

	instanceType = d.GetStringOrDefault("instanceType", instanceType)
	ami = d.GetStringOrDefault("ami", ami)
	ebsOptimized = d.GetBoolOrDefault("ebsOptimized", ebsOptimized)
	monitoring = d.GetBoolOrDefault("monitoring", monitoring)
	cpuCredits = d.GetStringOrDefault("creditSpecification.cpuCredits", cpuCredits)
	tenancy = d.GetStringOrDefault("tenancy", tenancy)

	a := &aws.Instance{
		Address:          d.Address,
		Region:           region.String(),
		Tenancy:          tenancy,
		PurchaseOption:   purchaseOption,
		AMI:              ami,
		InstanceType:     instanceType,
		EBSOptimized:     ebsOptimized,
		EnableMonitoring: monitoring,
		CPUCredits:       cpuCredits,
	}

	a.RootBlockDevice = &aws.EBSVolume{
		Address: "root_block_device",
		Region:  region.String(),
		Type:    d.Get("rootBlockDevice.volumeType").String(),
		IOPS:    d.Get("rootBlockDevice.iops").Int(),
	}

	if d.Get("rootBlockDevice.volumeSize").Type != gjson.Null {
		a.RootBlockDevice.Size = intPtr(d.Get("rootBlockDevice.volumeSize").Int())
	}
	// the # is a gjson directive meaning that its entering an array.
	ebsBlockDeviceRef := d.References("ebsBlockDevice.#.volumeId")

	for i, data := range d.Get("ebsBlockDevices").Array() {
		deviceName := data.Get("deviceName").String()

		ltDevice := ltEBSBlockDevices[deviceName]
		if ltDevice == nil {
			ltDevice = &aws.EBSVolume{}
		}

		// Check if there's a reference for this EBS volume
		// If there is then we shouldn't add as a subresource since
		// the cost will be added against the volume resource.
		if len(ebsBlockDeviceRef) > i && ebsBlockDeviceRef[i].Get("id").String() == data.Get("volumeId").String() {
			delete(ltEBSBlockDevices, deviceName)

			continue
		}

		// Instance can override individual Launch Template's values based on device
		// name.
		volumeType := ltDevice.Type
		if v := data.Get("volumeType"); v.Exists() {
			volumeType = v.String()
		}

		volumeSize := ltDevice.Size
		if v := data.Get("volumeSize"); v.Exists() {
			volumeSize = intPtr(v.Int())
		}

		iops := ltDevice.IOPS
		if v := data.Get("iops"); v.Exists() {
			iops = v.Int()
		}

		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("ebs_block_device[%d]", i),
			Region:  region.String(),
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

	a.PopulateUsage(u)

	return a.BuildResource()

}
