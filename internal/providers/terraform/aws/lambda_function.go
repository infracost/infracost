package aws

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLambdaFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_lambda_function",
		CoreRFunc: NewLambdaFunction,
	}
}

func NewLambdaFunction(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	name := d.Get("function_name").String()
	memorySize := int64(128)
	architectures := string("x86_64")
	storageSize := int64(512)

	if d.Get("memory_size").Type != gjson.Null {
		memorySize = d.Get("memory_size").Int()
	}

	if len(d.Get("architectures").Array()) > 0 {
		architectures = d.Get("architectures.0").String()
	}

	if d.Get("ephemeral_storage").Type != gjson.Null {
		storageSize = d.Get("ephemeral_storage.0.size").Int()
	}

	return &aws.LambdaFunction{
		Address:      d.Address,
		Region:       region,
		Name:         name,
		MemorySize:   memorySize,
		Architecture: architectures,
		StorageSize:  storageSize,
	}
}
