package terraform

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"
	"text/template"

	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/sync"
)

func setMockAttributes(blockAtts map[string]map[string]string) hcl.SetAttributesFunc {
	count := map[string]int{}

	return func(b *hcl.Block) {
		if _, ok := b.HCLBlock.Body.(*hclsyntax.Body); ok {
			if b.Type() == "resource" || b.Type() == "data" {
				b.UniqueAttrs = map[string]*hcl2.Attribute{}

				fullName := b.FullName()
				if attrs, ok := blockAtts[fullName]; ok {
					addAttrs(attrs, b)
				}

				withCount := fullName + "[0]"
				if i, ok := count[fullName]; ok {
					withCount = fmt.Sprintf("%s[%d]", fullName, i)
					count[fullName]++
				} else {
					count[fullName] = 0
				}

				if attrs, ok := blockAtts[withCount]; ok {
					addAttrs(attrs, b)
				}
			}

		}
	}
}

func addAttrs(attrs map[string]string, b *hcl.Block) {
	for k, v := range attrs {
		b.UniqueAttrs[k] = &hcl2.Attribute{
			Name: k,
			Expr: &hclsyntax.LiteralValueExpr{
				Val: cty.StringVal(v),
			},
		}
	}
}

func TestHCLProvider_LoadPlanJSON(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]map[string]string
		warnings []int
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
		{
			name: "populates warnings on missing vars",
			attrs: map[string]map[string]string{
				"aws_eip.eip": {
					"id":  "eip",
					"arn": "eip-arn",
				},
			},
			warnings: []int{
				105,
			},
		},
		{
			name: "shows correct duplicate variable warning",
		},
		{
			name: "builds module configuration correctly with count",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathName := strings.ReplaceAll(strings.ToLower(tt.name), " ", "_")
			testPath := path.Join("testdata/hcl_provider_test", pathName)

			ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{})
			moduleParser := modules.NewSharedHCLParser()
			pl := hcl.NewProjectLocator(nil)
			mods := pl.FindRootModules(testPath)
			options := []hcl.Option{hcl.OptionWithBlockBuilder(
				hcl.BlockBuilder{
					MockFunc: func(a *hcl.Attribute) cty.Value {
						return cty.StringVal(fmt.Sprintf("mocked-%s", a.Name()))
					},
					SetAttributes: []hcl.SetAttributesFunc{setMockAttributes(tt.attrs)},
					HCLParser:     moduleParser,
				},
			)}

			if mods[0].TerraformVarFiles != nil {
				options = append(options, hcl.OptionWithTFVarsPaths(mods[0].TerraformVarFiles.ToPaths(), true))
			}

			parser := hcl.NewParser(
				mods[0],
				hcl.CreateEnvFileMatcher([]string{}, nil),
				modules.NewModuleLoader(testPath, moduleParser, nil, config.TerraformSourceMap{}, &sync.KeyMutex{}),
				options...,
			)

			p := HCLProvider{
				Parser: parser,
				ctx:    ctx,
			}
			root := p.LoadPlanJSON()

			require.NoError(t, root.Error)

			// uncomment and run `make test` to update the expectations
			// var prettyJSON bytes.Buffer
			// err = json2.Indent(&prettyJSON, root.JSON, "", "  ")
			// assert.NoError(t, err)
			// err = os.WriteFile(path.Join(testPath, "expected.json"), append(prettyJSON.Bytes(), "\n"...), 0600)
			// assert.NoError(t, err)

			tmpl, err := template.ParseFiles(path.Join(testPath, "expected.json"))
			require.NoError(t, err)

			exp := bytes.NewBuffer([]byte{})
			err = tmpl.Execute(exp, map[string]interface{}{"attrs": tt.attrs})
			require.NoError(t, err)

			expected := exp.String()
			actual := string(root.JSON)
			assert.JSONEq(t, expected, actual)

			codes := make([]int, len(root.Module.Warnings))
			for i, w := range root.Module.Warnings {
				codes[i] = w.Code
			}

			assert.Len(t, codes, len(tt.warnings), "unexpected warning length")
			assert.ElementsMatch(t, codes, tt.warnings)
		})
	}
}
