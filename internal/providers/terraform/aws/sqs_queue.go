package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSQSQueueRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sqs_queue",
		RFunc: NewSqsQueue,
	}
}
func NewSqsQueue(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SqsQueue{Address: strPtr(d.Address), FifoQueue: boolPtr(d.Get("fifo_queue").Bool()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
