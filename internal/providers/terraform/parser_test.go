package terraform

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
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
				Name:         "aws_cloudwatch_log_group.each_resource[\"0\"]",
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
				Name:         "aws_cloudwatch_log_group.each_resource[\"1\"]",
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
						"address":"aws_cloudwatch_log_group.each_resource[\"0\"]",
						"mode":"managed",
						"type":"aws_cloudwatch_log_group",
						"name":"each_resource",
						"index":"0",
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
						"address":"aws_cloudwatch_log_group.each_resource[\"1\"]",
						"mode":"managed",
						"type":"aws_cloudwatch_log_group",
						"name":"each_resource",
						"index":"1",
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
				"address":"aws_cloudwatch_log_group.each_resource[\"0\"]",
				"mode":"managed",
				"type":"aws_cloudwatch_log_group",
				"name":"each_resource",
				"index":"0",
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
				"address":"aws_cloudwatch_log_group.each_resource[\"1\"]",
				"mode":"managed",
				"type":"aws_cloudwatch_log_group",
				"name":"each_resource",
				"index":"1",
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
							"address":"aws_cloudwatch_log_group.each_resource",
							"mode":"managed",
							"type":"aws_cloudwatch_log_group",
							"name":"each_resource",
							"provider_config_key":"aws",
							"expressions": {
								"name": {
									"references": [
										"each.key"
									]
								}
							},
							"schema_version":0,
							"for_each_expression": {
								"references": [
									"0",
									"1"
								]
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

	usage := schema.NewUsageMapFromInterface(map[string]any{
		"aws_cloudwatch_log_group.array_resource[*]": map[string]any{
			"monthly_data_ingested_gb": 0,
		},
		"aws_cloudwatch_log_group.each_resource[*]": map[string]any{
			"monthly_data_ingested_gb": 0,
		},
	})

	providerConf := parsed.Get("configuration.provider_config")
	conf := parsed.Get("configuration.root_module")
	vars := parsed.Get("variables")

	p := NewParser(config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, map[string]any{}), true)

	parsedResources := p.parseJSONResources(false, nil, usage, NewConfLoader(conf), parsed, providerConf, vars)
	actual := make([]*schema.Resource, len(parsedResources))
	for i, pr := range parsedResources {
		actual[i] = schema.BuildResource(pr.PartialResource, nil)
	}

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

	p := NewParser(config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, map[string]any{}), true)

	for _, test := range tests {
		parsed := p.createParsedResource(test.data, nil)
		actual := schema.BuildResource(parsed.PartialResource, nil)
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
          "values": {},
          "region": "eu-west-2"
				},
				{
					"address": "aws_instance.instance2",
					"mode": "managed",
          "type": "aws_instance",
          "name": "instance2",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {},
          "region": "eu-west-2"
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
							"values": {},
							"region": "eu-west-2"
						},
						{
							"address": "module.module1.aws_nat_gateway.nat2",
							"mode": "managed",
							"type": "aws_nat_gateway",
							"name": "nat2",
							"provider_name": "registry.terraform.io/hashicorp/aws",
							"schema_version": 0,
							"values": {},
							"region": "eu-west-2"
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
								"provider_config_key": "aws",
								"expressions": {}
							},
              {
								"address": "aws_nat_gateway.nat2",
								"mode": "managed",
								"type": "aws_nat_gateway",
								"name": "nat2",
								"provider_config_key": "aws.europe",
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

	p := NewParser(config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, map[string]any{}), true)
	loader := NewConfLoader(conf)
	actual := p.parseResourceData(false, loader, providerConf, planVals, vars)

	for k, v := range actual {
		assert.Equal(t, expected[k].Address, v.Address)
		assert.Equal(t, expected[k].ProviderName, v.ProviderName)
		assert.Equal(t, expected[k].Type, v.Type)

		region := p.getRegion(loader, v, providerConf, vars)
		assert.Equal(t, expectedRegions[k], region)
	}
}

func TestParseReferences_plan(t *testing.T) {
	vol1 := schema.NewResourceData("aws_ebs_volume", "aws", "aws_ebs_volume.volume1", nil, gjson.Result{
		Type: gjson.JSON,
		Raw:  `{}`,
	})

	snap1 := schema.NewResourceData("aws_ebs_snapshot", "aws", "aws_ebs_snapshot.snapshot1", nil, gjson.Result{
		Type: gjson.JSON,
		Raw:  `{}`,
	})

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

	p := NewParser(config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, map[string]any{}), true)
	p.parseReferences(resData, NewConfLoader(conf))

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}

func TestParseReferences_state(t *testing.T) {
	vol1 := schema.NewResourceData("aws_ebs_volume", "aws", "aws_ebs_volume.volume1", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"id": "vol-12345"
			}`,
	})

	snap1 := schema.NewResourceData("aws_ebs_snapshot", "aws", "aws_ebs_snapshot.snapshot1", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"volume_id": "vol-12345"
			}`,
	})

	resData := map[string]*schema.ResourceData{
		vol1.Address:  vol1,
		snap1.Address: snap1,
	}

	conf := gjson.Result{}

	p := NewParser(config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, map[string]any{}), true)
	p.parseReferences(resData, NewConfLoader(conf))

	assert.Equal(t, []*schema.ResourceData{vol1}, resData["aws_ebs_snapshot.snapshot1"].References("volume_id"))
}

func TestFixKnownModuleRefIssues(t *testing.T) {
	bucket := schema.NewResourceData("aws_s3_bucket", "registry.terraform.io/hashicorp/aws", "module.bucket.aws_s3_bucket.this[0]", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"id": "hcl-bucket-id"
				"acl":"private",
				"bucket":"my-bucket"
	}`,
	})

	policy := schema.NewResourceData("aws_s3_bucket_policy", "registry.terraform.io/hashicorp/aws", "module.bucket.aws_s3_bucket_policy.this[0]", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"id": "hcl-policy-id",
				"bucket":"my-bucket",
				"policy":"{}"
	}`,
	})

	lifecycleConfiguration := schema.NewResourceData("aws_s3_bucket_lifecycle_configuration", "registry.terraform.io/hashicorp/aws", "aws_s3_bucket_lifecycle_configuration.bucket_lifecycle", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"bucket": "hcl-policy-id"
	         }`,
	})

	lifecycleConfiguration.AddReference("bucket", policy, []string{})

	resData := map[string]*schema.ResourceData{
		lifecycleConfiguration.Address: lifecycleConfiguration,
		policy.Address:                 policy,
		bucket.Address:                 bucket,
	}

	require.Equal(t, lifecycleConfiguration.Get("bucket").String(), "hcl-policy-id")
	assert.Equal(t, lifecycleConfiguration.References("bucket")[0].Type, "aws_s3_bucket_policy")
	fixKnownModuleRefIssues(resData)

	assert.Equal(t, lifecycleConfiguration.Get("bucket").String(), "hcl-bucket-id")
	assert.Equal(t, lifecycleConfiguration.References("bucket")[0].Get("id").String(), "hcl-bucket-id")
	assert.Equal(t, lifecycleConfiguration.References("bucket")[0].Type, "aws_s3_bucket")
}

func TestParseKnownModuleRefs(t *testing.T) {
	res := schema.NewResourceData("aws_autoscaling_group", "registry.terraform.io/hashicorp/aws", "module.worker_groups_launch_template.aws_autoscaling_group.workers_launch_template[0]", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"capacity_rebalance":false,
				"desired_capacity":6,
				"enabled_metrics":null,
				"force_delete":false,
				"force_delete_warm_pool":false,
				"health_check_grace_period":300,
				"initial_lifecycle_hook":[],
				"instance_refresh":[],
				"launch_configuration":null,
				"launch_template":[{}],
				"load_balancers":null,
				"max_instance_lifetime":0,
				"max_size":3,
				"metrics_granularity":"1Minute",
				"min_elb_capacity":null,
				"min_size":1,
				"mixed_instances_policy":[],
				"name_prefix":"my-cluster-0",
				"placement_group":null,
				"protect_from_scale_in":false,
				"region":"us-east-1",
				"suspended_processes":["AZRebalance"],
				"tag":[
					{"key":"Name",
					"propagate_at_launch":true,
					"value":"my-cluster-0-eks_asg"},
					{"key":"kubernetes.io/cluster/my-cluster",
					"propagate_at_launch":true,
					"value":"owned"}
				],
				"tags":null,
				"target_group_arns":null,
				"termination_policies":[],
				"timeouts":null,
				"wait_for_capacity_timeout":"10m",
				"wait_for_elb_capacity":null,
				"warm_pool":[]}`,
	})

	lt := schema.NewResourceData("aws_launch_template", "registry.terraform.io/hashicorp/aws", "module.worker_groups_launch_template.aws_launch_template.workers_launch_template[0]", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{
				"block_device_mappings":[
					{
						"device_name":"/dev/xvda",
						"ebs":[
							{
								"delete_on_termination":"true",
								"encrypted":"false",
								"iops":0,
								"kms_key_id":"",
								"snapshot_id":null,
								"volume_size":100,
								"volume_type":"gp2"
							}
						],
					"no_device":null,
					"virtual_name":null
					}
				],
				"capacity_reservation_specification":[],
				"cpu_options":[],
				"credit_specification":[
					{
						"cpu_credits":"standard"
					}
				],
				"description":null,
				"disable_api_termination":null,
				"ebs_optimized":"true",
				"elastic_gpu_specifications":[],
				"elastic_inference_accelerator":[],
				"enclave_options":[{"enabled":false}],
				"hibernation_options":[],
				"iam_instance_profile":[{"arn":null}],
				"image_id":"ami-083ae86883eb12ef6",
				"instance_initiated_shutdown_behavior":null,
				"instance_market_options":[],
				"instance_type":"m4.large",
				"kernel_id":null,
				"key_name":"",
				"license_specification":[],
				"metadata_options":[
					{
						"http_endpoint":"enabled",
						"http_tokens":"optional"
					}
				],
				"monitoring":[{"enabled":true}],
				"name_prefix":"my-cluster-0",
				"network_interfaces":[
					{
						"associate_carrier_ip_address":null,
						"associate_public_ip_address":"false",
						"delete_on_termination":"true",
						"description":null,
						"device_index":null,
						"interface_type":null,
						"ipv4_address_count":null,
						"ipv4_addresses":null,
						"ipv6_address_count":null,
						"ipv6_addresses":null,
						"network_interface_id":null,
						"private_ip_address":null,
						"subnet_id":null
					}
				],
				"placement":[],
				"ram_disk_id":null,
				"region":"us-east-1",
				"security_group_names":null,
				"tag_specifications":[
					{
						"resource_type":"volume",
						"tags":
						{
							"Name":"my-cluster-0-eks_asg"
						}
					},
					{
						"resource_type":"instance",
						"tags":
						{
							"Name":"my-cluster-0-eks_asg"
						}
					}
				],
				"tags":null,
				"update_default_version":false,
				"vpc_security_group_ids":null
			}`,
	})

	resData := map[string]*schema.ResourceData{
		res.Address: res,
		lt.Address:  lt,
	}

	conf := gjson.Result{
		Type: gjson.JSON,
		Raw: `
		{
			"module_calls": {
				"worker_groups_launch_template": {
					"source": "terraform-aws-modules/eks/aws"
				}
			}
		}`,
	}
	assert.Nil(t, resData[res.Address].References("launch_template"))

	parseKnownModuleRefs(resData, NewConfLoader(conf))

	assert.NotNil(t, resData[res.Address].References("launch_template"))
}

func TestAddressResourcePart(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"aws_instance.my_instance", "aws_instance.my_instance"},
		{"data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance[\"index.1\"]"},
		{"data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance[\"index.1\"]"},
		// Modules
		{"module.my_module.aws_instance.my_instance", "aws_instance.my_instance"},
		{"module.my_module.data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"module.my_module.aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance[\"index.1\"]"},
		{"module.my_module.data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance[\"index.1\"]"},
		// Submodules
		{"module.my_module.module.my_submodule.aws_instance.my_instance", "aws_instance.my_instance"},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"module.my_module.module.my_submodule.aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance[\"index.1\"]"},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance[\"index.1\"]"},
		// Submodules with index
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance", "aws_instance.my_instance"},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance[\"index.1\"]"},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance[\"index.1\"]"},
	}

	for _, test := range tests {
		actual := addressResourcePart(test.address)
		assert.Equal(t, test.expected, actual)
	}
}

func TestAddressModulePart(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"aws_instance.my_instance", ""},
		{"data.aws_instance.my_instance", ""},
		{"aws_instance.my_instance[\"index.1\"]", ""},
		{"data.aws_instance.my_instance[\"index.1\"]", ""},
		// Modules
		{"module.my_module.aws_instance.my_instance", "module.my_module."},
		{"module.my_module.data.aws_instance.my_instance", "module.my_module."},
		{"module.my_module.aws_instance.my_instance[\"index.1\"]", "module.my_module."},
		{"module.my_module.data.aws_instance.my_instance[\"index.1\"]", "module.my_module."},
		// Submodules
		{"module.my_module.module.my_submodule.aws_instance.my_instance", "module.my_module.module.my_submodule."},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance", "module.my_module.module.my_submodule."},
		{"module.my_module.module.my_submodule.aws_instance.my_instance[\"index.1\"]", "module.my_module.module.my_submodule."},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance[\"index.1\"]", "module.my_module.module.my_submodule."},
		// Submodules with index
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance", "module.my_module.module.my_submodule[\"index.1\"]."},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance", "module.my_module.module.my_submodule[\"index.1\"]."},
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance[\"index.1\"]", "module.my_module.module.my_submodule[\"index.1\"]."},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance[\"index.1\"]", "module.my_module.module.my_submodule[\"index.1\"]."},
	}

	for _, test := range tests {
		actual := addressModulePart(test.address)
		assert.Equal(t, test.expected, actual)
	}
}

func TestRemoveAddressCountIndex(t *testing.T) {
	tests := []struct {
		address  string
		expected int
	}{
		{"aws_instance.my_instance", -1},
		{"data.aws_instance.my_instance", -1},
		{"aws_instance.my_instance[3]", 3},
		{"data.aws_instance.my_instance[3]", 3},
		{"aws_instance.my_instance[3]", 3},
		{"data.aws_instance.my_instance[3]", 3},
		// Modules
		{"module.my_module.aws_instance.my_instance", -1},
		{"module.my_module.data.aws_instance.my_instance", -1},
		{"module.my_module.aws_instance.my_instance[3]", 3},
		{"module.my_module.data.aws_instance.my_instance[3]", 3},
		{"module.my_module.aws_instance.my_instance[3]", 3},
		{"module.my_module.data.aws_instance.my_instance[3]", 3},
		// Count modules
		{"module.my_module[2].aws_instance.my_instance", -1},
		{"module.my_module[2].data.aws_instance.my_instance", -1},
		{"module.my_module[2].aws_instance.my_instance[3]", 3},
		{"module.my_module[2].data.aws_instance.my_instance[3]", 3},
		{"module.my_module[2].aws_instance.my_instance[3]", 3},
		{"module.my_module[2].data.aws_instance.my_instance[3]", 3},
		// Each resources
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance", -1},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance", -1},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index.1\"]", -1},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index.1\"]", -1},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index[1]\"]", -1},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index[1]\"]", -1},
	}

	for _, test := range tests {
		actual := addressCountIndex(test.address)
		assert.Equal(t, test.expected, actual)
	}
}

func TestAddressKey(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"aws_instance.my_instance", ""},
		{"data.aws_instance.my_instance", ""},
		{"aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"data.aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"aws_instance.my_instance[\"index[1]\"]", "index[1]"},
		{"data.aws_instance.my_instance[\"index[1]\"]", "index[1]"},
		// Modules
		{"module.my_module.aws_instance.my_instance", ""},
		{"module.my_module.data.aws_instance.my_instance", ""},
		{"module.my_module.aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"module.my_module.data.aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"module.my_module.aws_instance.my_instance[\"index[1]\"]", "index[1]"},
		{"module.my_module.data.aws_instance.my_instance[\"index[1]\"]", "index[1]"},
		// Each modules
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance", ""},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance", ""},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index.1\"]", "index.1"},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index[1]\"]", "index[1]"},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index[1]\"]", "index[1]"},
	}

	for _, test := range tests {
		actual := addressKey(test.address)
		assert.Equal(t, test.expected, actual)
	}
}

func TestRemoveAddressArrayPart(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"aws_instance.my_instance", "aws_instance.my_instance"},
		{"data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance"},
		{"data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance"},
		{"aws_instance.my_instance[\"index[1]\"]", "aws_instance.my_instance"},
		{"data.aws_instance.my_instance[\"index[1]\"]", "data.aws_instance.my_instance"},
		// Modules
		{"module.my_module.aws_instance.my_instance", "aws_instance.my_instance"},
		{"module.my_module.data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"module.my_module.aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance"},
		{"module.my_module.data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance"},
		{"module.my_module.aws_instance.my_instance[\"index[1]\"]", "aws_instance.my_instance"},
		{"module.my_module.data.aws_instance.my_instance[\"index[1]\"]", "data.aws_instance.my_instance"},
		// Each modules
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance", "aws_instance.my_instance"},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance", "data.aws_instance.my_instance"},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index.1\"]", "aws_instance.my_instance"},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index.1\"]", "data.aws_instance.my_instance"},
		{"module.my_module[\"modindex.0\"].aws_instance.my_instance[\"index[1]\"]", "aws_instance.my_instance"},
		{"module.my_module[\"modindex.0\"].data.aws_instance.my_instance[\"index[1]\"]", "data.aws_instance.my_instance"},
	}

	for _, test := range tests {
		actual := removeAddressArrayPart(test.address)
		assert.Equal(t, test.expected, actual)
	}
}

func TestGetModuleNames(t *testing.T) {
	tests := []struct {
		address  string
		expected []string
	}{
		{"aws_instance.my_instance", []string{}},
		{"data.aws_instance.my_instance", []string{}},
		{"aws_instance.my_instance[\"index.1\"]", []string{}},
		{"data.aws_instance.my_instance[\"index.1\"]", []string{}},
		// Modules
		{"module.my_module.aws_instance.my_instance", []string{"my_module"}},
		{"module.my_module.data.aws_instance.my_instance", []string{"my_module"}},
		{"module.my_module.aws_instance.my_instance[\"index.1\"]", []string{"my_module"}},
		{"module.my_module.data.aws_instance.my_instance[\"index.1\"]", []string{"my_module"}},
		// Submodules
		{"module.my_module.module.my_submodule.aws_instance.my_instance", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule.aws_instance.my_instance[\"index.1\"]", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule.data.aws_instance.my_instance[\"index.1\"]", []string{"my_module", "my_submodule"}},
		// Submodules with index
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule[\"index.1\"].aws_instance.my_instance[\"index.1\"]", []string{"my_module", "my_submodule"}},
		{"module.my_module.module.my_submodule[\"index.1\"].data.aws_instance.my_instance[\"index.1\"]", []string{"my_module", "my_submodule"}},
	}

	for _, test := range tests {
		actual := getModuleNames(test.address)
		assert.Equal(t, test.expected, actual)
	}
}
