package aws

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisFirehoseDeliveryStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_kinesis_firehose_delivery_stream",
		CoreRFunc: NewKinesisFirehoseDeliveryStream,
		ReferenceAttributes: []string{
			"elasticsearch_configuration.0.vpc_config.0.subnet_ids",
		},
	}
}

func NewKinesisFirehoseDeliveryStream(d *schema.ResourceData) schema.CoreResource {
	formatConversionEnabled := d.GetBoolOrDefault("extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", true)

	subnetIDs := len(d.Get("elasticsearch_configuration.0.vpc_config.0.subnet_ids").Array())

	// if the length of the subnet_ids attribute is zero this means that the attribute
	// has been modified with a subnet id that is yet to exist. In this instance we'll
	// use the reference attribute instead. In most cases this should have the accurate
	// number of subnet_ids.
	if subnetIDs == 0 {
		subnetIDs = len(d.References("elasticsearch_configuration.0.vpc_config.0.subnet_ids"))
	}

	r := &aws.KinesisFirehoseDeliveryStream{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		DataFormatConversionEnabled: d.Get("extended_s3_configuration.0.data_format_conversion_configuration").Exists() && formatConversionEnabled,
		VPCDeliveryEnabled:          d.Get("elasticsearch_configuration.0.vpc_config").Type != gjson.Null,
		VPCDeliveryAZs:              int64(subnetIDs),
	}
	return r
}
