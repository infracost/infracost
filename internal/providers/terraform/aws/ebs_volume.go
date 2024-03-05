package aws

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEBSVolumeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ebs_volume",
		CoreRFunc: NewEBSVolume,
	}
}

func NewEBSVolume(d *schema.ResourceData) schema.CoreResource {
	var size *int64
	if d.Get("size").Type != gjson.Null {
		size = intPtr(d.Get("size").Int())
	}

	a := &aws.EBSVolume{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		Type:       d.Get("type").String(),
		IOPS:       d.Get("iops").Int(),
		Throughput: d.Get("throughput").Int(),
		Size:       size,
	}

	return a
}
