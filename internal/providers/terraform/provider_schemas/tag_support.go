package provider_schemas

import (
	_ "embed"
	"encoding/json"
)

//go:embed aws.tags.json
var awsTagsJson []byte

//go:embed aws.tags_all.json
var awsTagsAllJson []byte

//go:embed azurerm.tags.json
var azurermTagsJson []byte

//go:embed google.labels.json
var googleLabelsJson []byte

var AwsTagsSupport map[string]bool
var AwsTagsAllSupport map[string]bool
var AzureTagsSupport map[string]bool
var GoogleLabelsSupport map[string]bool

func init() {
	err := json.Unmarshal(awsTagsJson, &AwsTagsSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(awsTagsAllJson, &AwsTagsAllSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(azurermTagsJson, &AzureTagsSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(googleLabelsJson, &GoogleLabelsSupport)
	if err != nil {
		panic(err)
	}
}
