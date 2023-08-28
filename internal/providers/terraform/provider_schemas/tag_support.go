package provider_schemas

import (
	_ "embed"
	"encoding/json"
)

//go:embed aws.tags.json
var awsTagsJSON []byte

//go:embed aws.tags_all.json
var awsTagsAllJSON []byte

//go:embed aws.tag_block.json
var awsTagBlockJSON []byte

//go:embed azurerm.tags.json
var azurermTagsJSON []byte

//go:embed google.labels.json
var googleLabelsJSON []byte

var AWSTagsSupport map[string]bool
var AWSTagsAllSupport map[string]bool
var AWSTagBlockSupport map[string]bool
var AzureTagsSupport map[string]bool
var GoogleLabelsSupport map[string]bool

func init() {
	err := json.Unmarshal(awsTagsJSON, &AWSTagsSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(awsTagsAllJSON, &AWSTagsAllSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(awsTagBlockJSON, &AWSTagBlockSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(azurermTagsJSON, &AzureTagsSupport)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(googleLabelsJSON, &GoogleLabelsSupport)
	if err != nil {
		panic(err)
	}
}
