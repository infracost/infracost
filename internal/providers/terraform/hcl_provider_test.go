package terraform

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"testing"
	"text/template"

	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/hcl"
)

func setMockAttributes(blockAtts map[string]map[string]string) hcl.SetAttributesFunc {
	count := map[string]int{}

	return func(moduleBlock *hcl.Block, block *hcl2.Block) {
		if v, ok := block.Body.(*hclsyntax.Body); ok {
			body := *v
			nat := hclsyntax.Attributes{}
			for k, a := range body.Attributes {
				b := *a
				nat[k] = &b
			}

			body.Attributes = nat

			if block.Type == "resource" || block.Type == "data" {
				fullName := strings.Join(block.Labels, ".")
				module := moduleBlock.FullName()
				if module != "" {
					fullName = module + "." + fullName
				}

				if attrs, ok := blockAtts[fullName]; ok {
					addAttrs(attrs, &body)
					block.Body = &body
				}

				withCount := fullName + "[0]"
				if i, ok := count[fullName]; ok {
					withCount = fmt.Sprintf("%s[%d]", fullName, i)
					count[fullName]++
				} else {
					count[fullName] = 0
				}

				if attrs, ok := blockAtts[withCount]; ok {
					addAttrs(attrs, &body)
					block.Body = &body
				}
			}
		}
	}
}

func addAttrs(attrs map[string]string, body *hclsyntax.Body) {
	for k, v := range attrs {
		body.Attributes[k] = &hclsyntax.Attribute{
			Name: k,
			Expr: &hclsyntax.LiteralValueExpr{
				Val: cty.StringVal(v),
			},
		}
	}
}

func TestHCLProvider_LoadPlanJSON(t *testing.T) {
	tests := []struct {
		name  string
		attrs map[string]map[string]string
	}{
		{
			name: "structures module expressions correctly with count",
			attrs: map[string]map[string]string{
				"module.module1.aws_ecs_task_definition.ecs_task": {
					"id":  "task-1",
					"arn": "task-1-arn",
				},
				"module.module1.module.module2.aws_ecs_task_definition.ecs_task": {
					"id":  "task-2",
					"arn": "task-2-arn",
				},
				"module.module1.aws_ecs_service.ecs_service": {
					"id":  "svc-1",
					"arn": "svc-1-arn",
				},
				"module.module1.module.module2.aws_ecs_service.ecs_service": {
					"id":  "svc-2",
					"arn": "svc-2-arn",
				},
			},
		},
		{
			name: "renders multiple count resources correctly",
			attrs: map[string]map[string]string{
				"aws_eip.test[0]": {
					"id":  "eip",
					"arn": "eip-arn",
				},
				"aws_eip.test[1]": {
					"id":  "eip-1",
					"arn": "eip-1-arn",
				},
				"module.autos.aws_autoscaling_group.test[0]": {
					"id":  "auto",
					"arn": "auto-arn",
				},
				"module.autos.aws_autoscaling_group.test[1]": {
					"id":  "auto-1",
					"arn": "auto-1-arn",
				},
				"module.autos.aws_autoscaling_group.test[2]": {
					"id":  "auto-2",
					"arn": "auto-2-arn",
				},
				"module.autos.aws_launch_configuration.test[0]": {
					"id":  "lc",
					"arn": "lc-arn",
				},
				"module.autos.aws_launch_configuration.test[1]": {
					"id":  "lc-1",
					"arn": "lc-1-arn",
				},
				"module.autos.aws_launch_configuration.test[2]": {
					"id":  "lc-2",
					"arn": "lc-2-arn",
				},
			},
		},
		{
			name: "renders module resources",
			attrs: map[string]map[string]string{
				"aws_vpn_connection.example": {
					"id":  "vpn-id",
					"arn": "vpn-arn",
				},
				"module.gateway.aws_customer_gateway.example": {
					"id":  "c-gw-id",
					"arn": "c-gw-arn",
				},
				"module.gateway.aws_ec2_transit_gateway.example": {
					"id":  "t-gw-id",
					"arn": "t-gw-arn",
				},
			},
		},
		{
			name: "does not panic on double attribute definition",
			attrs: map[string]map[string]string{
				"aws_eip.invalid_eip": {
					"id":  "eip",
					"arn": "eip-arn",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathName := strings.ReplaceAll(strings.ToLower(tt.name), " ", "_")
			testPath := path.Join("testdata/hcl_provider_test", pathName)

			logger := logrus.New()
			logger.SetOutput(io.Discard)
			parsers, err := hcl.LoadParsers(
				testPath,
				[]string{},
				logrus.NewEntry(logger),
				hcl.OptionWithBlockBuilder(
					hcl.BlockBuilder{
						MockFunc: func(a *hcl.Attribute) cty.Value {
							return cty.StringVal(fmt.Sprintf("mocked-%s", a.Name()))
						},
						SetAttributes: []hcl.SetAttributesFunc{setMockAttributes(tt.attrs)},
						Logger:        logrus.NewEntry(logger),
					},
				),
			)
			require.NoError(t, err)

			p := HCLProvider{
				parsers: parsers,
				logger:  logrus.NewEntry(logger),
			}
			got, err := p.LoadPlanJSONs()
			require.NoError(t, err)

			tmpl, err := template.ParseFiles(path.Join(testPath, "expected.json"))
			require.NoError(t, err)

			exp := bytes.NewBuffer([]byte{})
			err = tmpl.Execute(exp, map[string]interface{}{"attrs": tt.attrs})
			require.NoError(t, err)

			expected := exp.String()
			actual := string(got[0].json)
			assert.JSONEq(t, expected, actual)
		})
	}
}
