package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"

	"github.com/tidwall/gjson"
)

func getInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_instance",
		Notes: []string{
			"Costs associated with marketplace AMIs are not supported.",
			"For non-standard Linux AMIs such as Windows and RHEL, the operating system should be specified in usage file.",
			"EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.",
			"If a root volume is not specified then an 8Gi gp2 volume is assumed.",
		},
		RFunc:               NewInstance,
		ReferenceAttributes: []string{"ebs_block_device.#.volume_id"},
	}
}

func NewInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	a := &aws.Instance{
		Address:          d.Address,
		Region:           region,
		Tenancy:          d.Get("tenancy").String(),
		PurchaseOption:   purchaseOption,
		AMI:              d.Get("ami").String(),
		InstanceType:     d.Get("instance_type").String(),
		EBSOptimized:     d.Get("ebs_optimized").Bool(),
		EnableMonitoring: d.Get("monitoring").Bool(),
		CPUCredits:       d.Get("credit_specification.0.cpu_credits").String(),
	}

	a.RootBlockDevice = &aws.EBSVolume{
		Address: "root_block_device",
		Region:  region,
		Type:    d.Get("root_block_device.0.volume_type").String(),
		IOPS:    d.Get("root_block_device.0.iops").Int(),
	}

	if d.Get("root_block_device.0.volume_size").Type != gjson.Null {
		a.RootBlockDevice.Size = intPtr(d.Get("root_block_device.0.volume_size").Int())
	}

	ebsBlockDeviceRef := d.References("ebs_block_device.#.volume_id")

	for i, data := range d.Get("ebs_block_device").Array() {
		// Check if there's a reference for this EBS volume
		// If there is then we shouldn't add as a subresource since
		// the cost will be added against the volume resource.
		if len(ebsBlockDeviceRef) > i && ebsBlockDeviceRef[i].Get("id").String() == data.Get("volume_id").String() {
			continue
		}

		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("ebs_block_device[%d]", i),
			Region:  region,
			Type:    data.Get("volume_type").String(),
			IOPS:    data.Get("iops").Int(),
		}

		if data.Get("volume_size").Type != gjson.Null {
			ebsBlockDevice.Size = intPtr(data.Get("volume_size").Int())
		}

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	a.PopulateUsage(u)

	return a.BuildResource()

}
