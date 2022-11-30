package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var defaultEKSInstanceType = "t3.medium"

func getNewEKSNodeGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_eks_node_group",
		CoreRFunc: NewEKSNodeGroup,
		ReferenceAttributes: []string{
			"launch_template.0.id",
			"launch_template.0.name",
		},
	}
}

func NewEKSNodeGroup(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()

	instanceCount := d.Get("scaling_config.0.desired_size").Int()

	diskSize := int64(20)
	if d.Get("disk_size").Exists() {
		diskSize = d.Get("disk_size").Int()
	}
	a := &aws.EKSNodeGroup{
		Address:       d.Address,
		Region:        region,
		Name:          d.Get("node_group_name").String(),
		ClusterName:   d.Get("cluster_name").String(),
		InstanceCount: intPtr(instanceCount),
		DiskSize:      float64(diskSize),
	}

	launchTemplateRefID := d.References("launch_template.0.id")
	launchTemplateRefName := d.References("launch_template.0.name")
	launchTemplateRef := []*schema.ResourceData{}

	if len(launchTemplateRefID) > 0 {
		launchTemplateRef = launchTemplateRefID
	} else if len(launchTemplateRefName) > 0 {
		launchTemplateRef = launchTemplateRefName
	}

	// The instance types in the eks_node_group resource overrides any in the launch template
	instanceType := strings.ToLower(d.Get("instance_types.0").String())

	if len(launchTemplateRef) > 0 {
		data := launchTemplateRef[0]

		onDemandPercentageAboveBaseCount := int64(100)
		if strings.ToLower(launchTemplateRef[0].Get("instance_market_options.0.market_type").String()) == "spot" {
			onDemandPercentageAboveBaseCount = int64(0)
		}

		if instanceType != "" {
			data.Set("instance_type", instanceType)
		}

		if data.IsEmpty("instance_type") {
			data.Set("instance_type", defaultEKSInstanceType)
		}

		a.LaunchTemplate = newLaunchTemplate(data, region, instanceCount, int64(0), onDemandPercentageAboveBaseCount)
	} else {
		if instanceType == "" {
			instanceType = defaultEKSInstanceType
		}

		a.InstanceType = instanceType
		a.PurchaseOption = strings.ToLower(d.Get("capacity_type").String())
	}

	return a
}
