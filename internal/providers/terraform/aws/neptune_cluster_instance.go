package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNeptuneClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_neptune_cluster_instance",
		CoreRFunc: NewNeptuneClusterInstance,
		ReferenceAttributes: []string{
			"cluster_identifier",
		},
	}
}

func NewNeptuneClusterInstance(d *schema.ResourceData) schema.CoreResource {
	ioOptimized := false

	clusterIdentifiers := d.References("cluster_identifier")
	if len(clusterIdentifiers) > 0 {
		cluster := clusterIdentifiers[0]
		ioOptimized = strings.EqualFold(cluster.Get("storage_type").String(), "iopt1")
	}

	r := &aws.NeptuneClusterInstance{
		Address:       d.Address,
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("instance_class").String(),
		IOOptimized:   ioOptimized,
	}

	if !d.IsEmpty("count") {
		r.Count = intPtr(d.Get("count").Int())
	}
	return r
}
