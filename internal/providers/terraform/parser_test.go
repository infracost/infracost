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

func TestCreateResource(t *testing.T) {
	tests := []struct {
		data     *schema.ResourceData
		expected *schema.Resource
	}{
		{
			data: &schema.ResourceData{
				Address: "aws_instance.supported_resource",
				Type:    "aws_instance",
			},
			expected: &schema.Resource{
				Name:         "aws_instance.supported_resource",
				ResourceType: "aws_instance",
				IsSkipped:    false,
				NoPrice:      false,
			},
		},
		{
			data: &schema.ResourceData{
				Address: "null_resource.free_resource",
				Type:    "null_resource",
			},
			expected: &schema.Resource{
				Name:         "null_resource.free_resource",
				ResourceType: "null_resource",
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
			},
		},
		{
			data: &schema.ResourceData{
				Address: "fake_resource.unsupported_resource",
				Type:    "fake_resource",
			},
			expected: &schema.Resource{
				Name:         "fake_resource.unsupported_resource",
				ResourceType: "fake_resource",
				IsSkipped:    true,
				NoPrice:      false,
				SkipMessage:  "This resource is not currently supported",
			},
		},
	}

	p := NewParser(config.NewEnvironment())

	for _, test := range tests {
		actual := p.createResource(test.data, nil)
		assert.Equal(t, test.expected.Name, actual.Name)
		assert.Equal(t, test.expected.ResourceType, actual.ResourceType)
		assert.Equal(t, test.expected.IsSkipped, actual.IsSkipped)
		assert.Equal(t, test.expected.SkipMessage, actual.SkipMessage)
	}
}

func TestParseResourceData(t *testing.T) {
	providerConf := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
			"aws": {
				"name": "aws",
				"expressions": {
					"region": {
						"constant_value": "us-west-2"
					}
				}
			},
			"aws.europe": {
				"name": "aws",
				"alias": "europe",
				"expressions": {
					"region": {
						"references": ["var.reg_var"]
					}
				}
			},
      "module.module1:aws.europe": {
        "name": "aws",
        "alias": "europe",
        "module_address": "module.module1"
      },
		}`,
	}

	planVals := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
			"resources": [
				{
					"address": "aws_instance.instance1",
					"mode": "managed",
          "type": "aws_instance",
          "name": "instance1",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {}
				},
				{
					"address": "aws_instance.instance2",
					"mode": "managed",
          "type": "aws_instance",
          "name": "instance2",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {}
        }
			],
			"child_modules": [
				{
					"address": "module.module1",
					"resources": [
						{
							"address": "module.module1.aws_nat_gateway.nat1",
							"mode": "managed",
							"type": "aws_nat_gateway",
							"name": "nat1",
							"provider_name": "registry.terraform.io/hashicorp/aws",
							"schema_version": 0,
							"values": {}
						},
						{
							"address": "module.module1.aws_nat_gateway.nat2",
							"mode": "managed",
							"type": "aws_nat_gateway",
							"name": "nat2",
							"provider_name": "registry.terraform.io/hashicorp/aws",
							"schema_version": 0,
							"values": {}
						}
					]
				}
			]
		}`,
	}

	conf := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
			"resources": [
				{
					"address": "aws_instance.instance1",
					"mode": "managed",
          "type": "aws_instance",
          "name": "instance1",
          "provider_config_key": "aws",
          "expressions": {}
				},
				{
					"address": "aws_instance.instance2",
					"mode": "managed",
          "type": "aws_instance",
          "name": "instance2",
          "provider_config_key": "aws.europe",
          "expressions": {}
				}
			],
      "module_calls": {
        "module1": {
          "source": "./module1",
          "module": {
            "resources": [
              {
								"address": "aws_nat_gateway.nat1",
								"mode": "managed",
								"type": "aws_nat_gateway",
								"name": "nat1",
								"provider_config_key": "module1:aws",
								"expressions": {}
							},
              {
								"address": "aws_nat_gateway.nat2",
								"mode": "managed",
								"type": "aws_nat_gateway",
								"name": "nat2",
								"provider_config_key": "module1:aws.europe",
								"expressions": {}
							}
						]
					}
				}
			}
		}`,
	}

	vars := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"reg_var": {
					"value": "eu-west-2"
				}
			}`,
	}

	expected := map[string]*schema.ResourceData{
		"aws_instance.instance1": {
			Address:      "aws_instance.instance1",
			ProviderName: "registry.terraform.io/hashicorp/aws",
			Type:         "aws_instance",
		},
		"aws_instance.instance2": {
			Address:      "aws_instance.instance2",
			ProviderName: "registry.terraform.io/hashicorp/aws",
			Type:         "aws_instance",
		},
		"module.module1.aws_nat_gateway.nat1": {
			Address:      "module.module1.aws_nat_gateway.nat1",
			ProviderName: "registry.terraform.io/hashicorp/aws",
			Type:         "aws_nat_gateway",
		},
		"module.module1.aws_nat_gateway.nat2": {
			Address:      "module.module1.aws_nat_gateway.nat2",
			ProviderName: "registry.terraform.io/hashicorp/aws",
			Type:         "aws_nat_gateway",
		},
	}

	expectedRegions := map[string]string{
		"aws_instance.instance1":              "us-west-2",
		"aws_instance.instance2":              "eu-west-2",
		"module.module1.aws_nat_gateway.nat1": "us-west-2",
		"module.module1.aws_nat_gateway.nat2": "eu-west-2",
	}

	p := NewParser(config.NewEnvironment())
	actual := p.parseResourceData(providerConf, planVals, conf, vars)

	for k, v := range actual {
		assert.Equal(t, expected[k].Address, v.Address)
		assert.Equal(t, expected[k].ProviderName, v.ProviderName)
		assert.Equal(t, expected[k].Type, v.Type)
		assert.Equal(t, expectedRegions[k], v.Get("region").String())
	}
}

func TestParseReferences_plan(t *testing.T) {
	vol1 := schema.NewResourceData(
		"aws_ebs_volume",
		"aws",
		"aws_ebs_volume.volume1",
		map[string]string{},
		gjson.Result{
			Type: gjson.JSON,
			Raw:  `{}`,
		},
	)

	snap1 := schema.NewResourceData(
		"aws_ebs_snapshot",
		"aws",
		"aws_ebs_snapshot.snapshot1",
		map[string]string{},
		gjson.Result{
			Type: gjson.JSON,
			Raw:  `{}`,
		},
	)

	resData := map[string]*schema.ResourceData{
		vol1.Address:  vol1,
		snap1.Address: snap1,
	}

	conf := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
			"resources": [
				{
					"address": "aws_ebs_volume.volume1",
					"mode": "managed",
          "type": "aws_ebs_volume",
          "name": "volume1",
          "provider_config_key": "aws",
          "expressions": {}
				},
				{
					"address": "aws_ebs_snapshot.snapshot1",
					"mode": "managed",
          "type": "aws_ebs_snapshot",
          "name": "snapshot1",
          "provider_config_key": "aws",
          "expressions": {
            "volume_id": {
              "references": [
                "aws_ebs_volume.volume1"
              ]
            }
					}
				}
			],
		}`,
	}

	p := NewParser(config.NewEnvironment())
	p.parseReferences(resData, conf)

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}

func TestParseReferences_state(t *testing.T) {
	vol1 := schema.NewResourceData(
		"aws_ebs_volume",
		"aws",
		"aws_ebs_volume.volume1",
		map[string]string{},
		gjson.Result{
			Type: gjson.JSON,
			Raw: `{
				"id": "vol-12345"
			}`,
		},
	)

	snap1 := schema.NewResourceData(
		"aws_ebs_snapshot",
		"aws",
		"aws_ebs_snapshot.snapshot1",
		map[string]string{},
		gjson.Result{
			Type: gjson.JSON,
			Raw: `{
				"volume_id": "vol-12345"
			}`,
		},
	)

	resData := map[string]*schema.ResourceData{
		vol1.Address:  vol1,
		snap1.Address: snap1,
	}

	conf := gjson.Result{}

	p := NewParser(config.NewEnvironment())
	p.parseReferences(resData, conf)

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}
