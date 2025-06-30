package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var defaultEKSInstanceType = "t3.medium"

func getNewEKSNodeGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_node_group",
		RFunc: NewEKSNodeGroup,
		ReferenceAttributes: []string{
			"launchTemplate.0.id",
			"launchTemplate.0.name",
		},
	}
}

func NewEKSNodeGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	instanceCount := d.Get("scalingConfig.desiredSize").Int()

	diskSize := int64(20)
	if d.Get("diskSize").Exists() {
		diskSize = d.Get("diskSize").Int()
	}
	a := &aws.EKSNodeGroup{
		Address:       d.Address,
		Region:        region,
		Name:          d.Get("nodeGroupName").String(),
		ClusterName:   d.Get("clusterName").String(),
		InstanceCount: intPtr(instanceCount),
		DiskSize:      float64(diskSize),
	}

	launchTemplateRefID := d.References("launchTemplate.0.id")
	launchTemplateRefName := d.References("launchTemplate.0.name")
	launchTemplateRef := []*schema.ResourceData{}

	if len(launchTemplateRefID) > 0 {
		launchTemplateRef = launchTemplateRefID
	} else if len(launchTemplateRefName) > 0 {
		launchTemplateRef = launchTemplateRefName
	}

	// The instance types in the eks_node_group resource overrides any in the launch template
	instanceType := strings.ToLower(d.Get("instanceTypes.0").String())

	if len(launchTemplateRef) > 0 {
		data := launchTemplateRef[0]

		onDemandPercentageAboveBaseCount := int64(100)
		if strings.ToLower(launchTemplateRef[0].Get("instanceMarketOptions.0.marketType").String()) == "spot" {
			onDemandPercentageAboveBaseCount = int64(0)
		}

		if instanceType != "" {
			data.Set("instanceType", instanceType)
		}

		if data.IsEmpty("instance_type") {
			data.Set("instanceType", defaultEKSInstanceType)
		}

		a.LaunchTemplate = newLaunchTemplate(data, u, region, instanceCount, int64(0), onDemandPercentageAboveBaseCount)
	} else {
		if instanceType == "" {
			instanceType = defaultEKSInstanceType
		}

		a.InstanceType = instanceType
		a.PurchaseOption = strings.ToLower(d.Get("capacityType").String())
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}
