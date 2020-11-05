package terraform

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

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

	for _, test := range tests {
		actual := createResource(test.data, nil)
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

	actual := parseResourceData(providerConf, planVals, conf, vars)

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
		gjson.Result{
			Type: gjson.JSON,
			Raw:  `{}`,
		},
	)

	snap1 := schema.NewResourceData(
		"aws_ebs_snapshot",
		"aws",
		"aws_ebs_snapshot.snapshot1",
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

	parseReferences(resData, conf)

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}

func TestParseReferences_state(t *testing.T) {
	vol1 := schema.NewResourceData(
		"aws_ebs_volume",
		"aws",
		"aws_ebs_volume.volume1",
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

	parseReferences(resData, conf)

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}
