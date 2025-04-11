package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_kinesis_stream",
		RFunc: newKinesisStream,
	}
}

func newKinesisStream(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	StreamMode := d.Get("streamModeDetails.0.streamMode").String()
	ShardCount := d.Get("shardCount").Int()

	return &aws.KinesisStream{
		Address:    d.Address,
		Region:     region,
		StreamMode: StreamMode,
		ShardCount: ShardCount,
	}
}
