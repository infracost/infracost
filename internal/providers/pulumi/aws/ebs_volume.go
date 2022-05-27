package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func getEBSVolumeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws:ebs/volume:Volume",
		RFunc: NewEBSVolume,
	}
}

func NewEBSVolume(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	log.Debugf("resources %s", d)
	var size *int64
	if d.Get("size").Type != gjson.Null {
		size = intPtr(d.Get("size").Int())
	}
	var region = d.Get("config.aws:region")
	log.Debugf("region %s", region)
	a := &aws.EBSVolume{
		Address:    d.Address,
		Region:     region.String(),
		Type:       d.Get("type").String(),
		IOPS:       d.Get("iops").Int(),
		Throughput: d.Get("throughput").Int(),
		Size:       size,
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}
