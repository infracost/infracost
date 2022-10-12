package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getLambdaFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_lambda_function",
		Notes:     []string{"Provisioned concurrency is not yet supported."},
		CoreRFunc: NewLambdaFunction,
	}
}

func NewLambdaFunction(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	name := d.Get("function_name").String()
	memorySize := int64(128)
	if d.Get("memory_size").Type != gjson.Null {
		memorySize = d.Get("memory_size").Int()
	}

	return &aws.LambdaFunction{
		Address:    d.Address,
		Region:     region,
		Name:       name,
		MemorySize: memorySize,
	}
}
