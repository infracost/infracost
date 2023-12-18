package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_kinesis_stream",
		CoreRFunc: newKinesisStream,
	}
}

func newKinesisStream(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	StreamMode := d.Get("stream_mode_details.0.stream_mode").String()
	ShardCount := d.Get("shard_count").Int()

	return &aws.KinesisStream{
		Address:    d.Address,
		Region:     region,
		StreamMode: StreamMode,
		ShardCount: ShardCount,
	}
}
