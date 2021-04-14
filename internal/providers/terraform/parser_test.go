package terraform

import (
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestParseJSONResources(t *testing.T) {
	var unknown *decimal.Decimal

	tests := []struct {
		expected *schema.Resource
	}{
		{
			expected: &schema.Resource{
				Name:         "aws_cloudwatch_log_group.array_resource[0]",
				ResourceType: "aws_cloudwatch_log_group",
				IsSkipped:    false,
				NoPrice:      false,
				CostComponents: []*schema.CostComponent{
					{
						Name:            "Data ingested",
						MonthlyQuantity: &decimal.Zero,
					},
				},
			},
		},
		{
			expected: &schema.Resource{
				Name:         "aws_cloudwatch_log_group.array_resource[1]",
				ResourceType: "aws_cloudwatch_log_group",
				IsSkipped:    false,
				NoPrice:      false,
				CostComponents: []*schema.CostComponent{
					{
						Name:            "Data ingested",
						MonthlyQuantity: &decimal.Zero,
					},
				},
			},
		},
		{
			expected: &schema.Resource{
				Name:         "aws_cloudwatch_log_group.non_array_resource",
				ResourceType: "aws_cloudwatch_log_group",
				IsSkipped:    false,
				NoPrice:      false,
				CostComponents: []*schema.CostComponent{
					{
						Name:            "Data ingested",
						MonthlyQuantity: unknown,
					},
				},
			},
		},
	}

	testData := `
	{
		"format_version":"0.1",
		"terraform_version":"0.14.8",
		"planned_values": {
			"root_module": {
				"resources": [
					{
						"address":"aws_cloudwatch_log_group.array_resource[0]",
						"mode":"managed",
						"type":"aws_cloudwatch_log_group",
						"name":"array_resource",
						"index":0,
						"provider_name":"registry.terraform.io/hashicorp/aws",
						"schema_version":0,
						"values": {
							"kms_key_id":null,
							"name":"log-group0",
							"name_prefix":null,
							"retention_in_days":0,
							"tags":null
						}
					},
					{
						"address":"aws_cloudwatch_log_group.array_resource[1]",
						"mode":"managed",
						"type":"aws_cloudwatch_log_group",
						"name":"array_resource",
						"index":1,
						"provider_name":"registry.terraform.io/hashicorp/aws",
						"schema_version":0,
						"values": {
							"kms_key_id":null,
							"name":"log-group1",
							"name_prefix":null,
							"retention_in_days":0,
							"tags":null
						}
					},
					{
						"address":"aws_cloudwatch_log_group.non_array_resource",
						"mode":"managed",
						"type":"aws_cloudwatch_log_group",
						"name":"non_array_resource",
						"provider_name":"registry.terraform.io/hashicorp/aws",
						"schema_version":0,
						"values": {
							"kms_key_id":null,
							"name":"log-group",
							"name_prefix":null,
							"retention_in_days":0,
							"tags":null
						}
					}
				]
			}
		},
		"resource_changes": [
			{
				"address":"aws_cloudwatch_log_group.array_resource[0]",
				"mode":"managed",
				"type":"aws_cloudwatch_log_group",
				"name":"array_resource",
				"index":0,
				"provider_name":"registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": [
						"create"
					],
					"before":null,
					"after": {
						"kms_key_id":null,
						"name":"log-group0",
						"name_prefix":null,
						"retention_in_days":0,
						"tags":null
					},
					"after_unknown": {
						"arn":true,
						"id":true
					}
				}
			},
			{
				"address":"aws_cloudwatch_log_group.array_resource[1]",
				"mode":"managed",
				"type":"aws_cloudwatch_log_group",
				"name":"array_resource",
				"index":1,
				"provider_name":"registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": [
						"create"
					],
					"before":null,
					"after": {
						"kms_key_id":null,
						"name":"log-group1",
						"name_prefix":null,
						"retention_in_days":0,
						"tags":null
					},
					"after_unknown": {
						"arn":true,
						"id":true
					}
				}
			},
			{
				"address":"aws_cloudwatch_log_group.non_array_resource",
				"mode":"managed",
				"type":"aws_cloudwatch_log_group",
				"name":"non_array_resource",
				"provider_name":"registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": [
						"create"
					],
					"before":null,
					"after": {
						"kms_key_id":null,
						"name":"log-group",
						"name_prefix":null,
						"retention_in_days":0,
						"tags":null
					},
					"after_unknown": {
						"arn":true,"id":true
					}
				}
			}
		],
		"configuration": {
			"provider_config": {
				"aws": {
					"name":"aws",
					"expressions": {
						"access_key": {
							"constant_value":"mock_access_key"
						},
						"region": {
							"constant_value":"us-east-1"
						},
						"secret_key": {
							"constant_value":"mock_secret_key"
						},
						"skip_credentials_validation": {
							"constant_value":true
						},
						"skip_requesting_account_id": {
							"constant_value":true
						}
					}
				},
				"google": {
					"name":"google",
					"expressions": {
						"credentials": {
							"constant_value":"{\"type\":\"service_account\"}"},
							"region": {
								"constant_value":"us-central1"
							}
						}
					}
				},
				"root_module": {
					"resources": [
						{
							"address":"aws_cloudwatch_log_group.array_resource",
							"mode":"managed",
							"type":"aws_cloudwatch_log_group",
							"name":"array_resource",
							"provider_config_key":"aws",
							"expressions": {
								"name": {
									"references": [
										"count.index"
									]
								}
							},
							"schema_version":0,
							"count_expression": {
								"constant_value":2
							}
						},
						{
							"address":"aws_cloudwatch_log_group.non_array_resource",
							"mode":"managed",
							"type":"aws_cloudwatch_log_group",
							"name":"non_array_resource",
							"provider_config_key":"aws",
							"expressions": {
								"name": {
									"constant_value":"log-group"
								}
							},
							"schema_version":0
						}
					]
				}
			}
		}
	}`

	parsed := gjson.Parse(testData)

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_cloudwatch_log_group.array_resource[*]": map[string]interface{}{
			"monthly_data_ingested_gb": 0,
		},
	})

	providerConf := parsed.Get("configuration.provider_config")
	conf := parsed.Get("configuration.root_module")
	vars := parsed.Get("variables")

	p := NewParser(config.NewEnvironment())

	actual := p.parseJSONResources(false, nil, usage, parsed, providerConf, conf, vars)

	i := 0
	for _, test := range tests {
		var resource *schema.Resource
		for _, act := range actual {
			if test.expected.Name == act.Name {
				resource = act
			}
		}

		assert.Equal(t, test.expected.CostComponents[0].Name, resource.CostComponents[0].Name)
		if test.expected.CostComponents[0].MonthlyQuantity != nil {
			assert.Equal(t, test.expected.CostComponents[0].MonthlyQuantity.BigInt(), resource.CostComponents[0].MonthlyQuantity.BigInt())
		} else {
			assert.Equal(t, test.expected.CostComponents[0].MonthlyQuantity, resource.CostComponents[0].MonthlyQuantity)
		}
		i++
	}
}
