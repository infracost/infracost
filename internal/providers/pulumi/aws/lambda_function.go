package aws

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLambdaFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lambda_function",
		RFunc: NewLambdaFunction,
	}
}

func NewLambdaFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	name := d.Get("functionName").String()
	memorySize := int64(128)
	architectures := string("x86_64")
	storageSize := int64(512)

	if d.Get("memorySize").Type != gjson.Null {
		memorySize = d.Get("memorySize").Int()
	}

	if len(d.Get("architectures").Array()) > 0 {
		architectures = d.Get("architectures.0").String()
	}

	if d.Get("ephemeralStorage").Type != gjson.Null {
		storageSize = d.Get("ephemeralStorage.size").Int()
	}

	a := &aws.LambdaFunction{
		Address:      d.Address,
		Region:       region,
		Name:         name,
		MemorySize:   memorySize,
		Architecture: architectures,
		StorageSize:  storageSize,
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}