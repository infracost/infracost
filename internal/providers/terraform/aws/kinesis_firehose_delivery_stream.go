package aws

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisFirehoseDeliveryStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesis_firehose_delivery_stream",
		RFunc: NewKinesisFirehoseDeliveryStream,
	}
}

func NewKinesisFirehoseDeliveryStream(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	formatConversionEnabled := d.GetBoolOrDefault("extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", true)
	r := &aws.KinesisFirehoseDeliveryStream{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		DataFormatConversionEnabled: d.Get("extended_s3_configuration.0.data_format_conversion_configuration").Exists() && formatConversionEnabled,
		VPCDeliveryEnabled:          d.Get("elasticsearch_configuration.0.vpc_config").Type != gjson.Null,
		VPCDeliveryAZs:              int64(len(d.Get("elasticsearch_configuration.0.vpc_config.0.subnet_ids").Array())),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
