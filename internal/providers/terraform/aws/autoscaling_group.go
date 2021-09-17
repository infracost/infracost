package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_autoscaling_group",
		RFunc: NewAutoscalingGroup,
		ReferenceAttributes: []string{
			"launch_configuration",
			"launch_template.0.id",
			"launch_template.0.name",
			"mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id",
			"launch_template",
		},
	}
}

func NewAutoscalingGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	a := &aws.AutoscalingGroup{
		Address:  d.Address,
		Capacity: d.Get("desired_capacity").Int(),
	}

	if len(d.References("launch_configuration")) > 0 {
		lc := d.References("launch_configuration")[0]
		if lc == nil {
			return nil
		}

		a.LaunchConfiguration = newLaunchConfiguration(lc, u, d.Get("region").String())
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}

func newLaunchConfiguration(d *schema.ResourceData, u *schema.UsageData, region string) *aws.LaunchConfiguration {
	tenancy := "Shared"
	if strings.ToLower(d.Get("placement_tenancy").String()) == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	} else if strings.ToLower(d.Get("placement_tenancy").String()) == "dedicated" {
		tenancy = "Dedicated"
	}

	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	a := &aws.LaunchConfiguration{
		Address:          d.Address,
		Region:           region,
		Tenancy:          tenancy,
		PurchaseOption:   purchaseOption,
		InstanceType:     d.Get("instance_type").String(),
		EBSOptimized:     d.Get("ebs_optimized").Bool(),
		EnableMonitoring: d.Get("enable_monitoring").Bool(),
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

	for i, data := range d.Get("ebs_block_device").Array() {
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

	return a
}
