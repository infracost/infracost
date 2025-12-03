package terraform

import (
	"bytes"
	json2 "encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"text/template"

	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/credentials"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/sync"
)

var update = flag.Bool("update", false, "update .golden files")

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
		chdir    bool
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
		{
			name: "adds_source_url_from_remote_module",
		},
		{
			name:  "adds_source_url_from_remote_module_chdir",
			chdir: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathName := strings.ReplaceAll(strings.ToLower(tt.name), " ", "_")
			testPath := path.Join("testdata/hcl_provider_test", pathName)

			logger := zerolog.New(io.Discard)

			ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, logrus.Fields{})
			moduleParser := modules.NewSharedHCLParser()
			pl := hcl.NewProjectLocator(logger, nil)
			startingPath := testPath
			initialPath, err := os.Getwd()
			require.NoError(t, err)
			if tt.chdir {
				err := os.Chdir(testPath)
				require.NoError(t, err)
				startingPath = "."
			}

			mods, _ := pl.FindRootModules(startingPath)
			options := []hcl.Option{hcl.OptionWithBlockBuilder(
				hcl.BlockBuilder{
					MockFunc: func(a *hcl.Attribute) cty.Value {
						return cty.StringVal(fmt.Sprintf("mocked-%s", a.Name()))
					},
					SetAttributes: []hcl.SetAttributesFunc{setMockAttributes(tt.attrs)},
					Logger:        logger,
					HCLParser:     moduleParser,
				},
			)}

			if mods[0].TerraformVarFiles != nil {
				options = append(options, hcl.OptionWithTFVarsPaths(mods[0].TerraformVarFiles.ToPaths(), true))
			}

			parser := hcl.NewParser(
				mods[0],
				hcl.CreateEnvFileMatcher([]string{}, nil),
				modules.NewModuleLoader(modules.ModuleLoaderOptions{
					CachePath:         startingPath,
					HCLParser:         moduleParser,
					CredentialsSource: &modules.CredentialsSource{FetchToken: credentials.FindTerraformCloudToken},
					SourceMap:         config.TerraformSourceMap{},
					SourceMapRegex:    nil,
					Logger:            logger,
					ModuleSync:        &sync.KeyMutex{},
				}),
				logger,
				options...,
			)

			p := HCLProvider{
				Parser: parser,
				logger: logger,
				ctx:    ctx,
			}
			root := p.LoadPlanJSON()

			require.NoError(t, root.Error)
			if tt.chdir {
				err = os.Chdir(initialPath)
				require.NoError(t, err)
			}

			// uncomment and run `make test` to update the expectations
			// var prettyJSON bytes.Buffer
			// err = json2.Indent(&prettyJSON, root.JSON, "", "  ")
			// assert.NoError(t, err)
			// err = os.WriteFile(path.Join(testPath, "expected.json"), append(prettyJSON.Bytes(), "\n"...), 0600)
			// assert.NoError(t, err)

			tmpl, err := template.ParseFiles(path.Join(testPath, "expected.json"))
			require.NoError(t, err)

			exp := bytes.NewBuffer([]byte{})
			err = tmpl.Execute(exp, map[string]any{"attrs": tt.attrs})
			require.NoError(t, err)

			expected := exp.String()
			actual := string(root.JSON)
			if !assert.JSONEq(t, expected, actual) {
				var prettyJSON bytes.Buffer
				err = json2.Indent(&prettyJSON, root.JSON, "", "  ")
				assert.NoError(t, err)

				if update != nil && *update {
					err = os.WriteFile(path.Join(testPath, "expected.json"), append(prettyJSON.Bytes(), "\n"...), 0600)
					assert.NoError(t, err)
				} else {
					err = os.WriteFile(path.Join(testPath, "actual.json"), append(prettyJSON.Bytes(), "\n"...), 0600)
					assert.NoError(t, err)
				}
			}

			codes := make([]int, len(root.Module.Warnings))
			for i, w := range root.Module.Warnings {
				codes[i] = w.Code
			}

			assert.Len(t, codes, len(tt.warnings), "unexpected warning length")
			assert.ElementsMatch(t, codes, tt.warnings)
		})
	}
}

func TestNormalizeModuleURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:user/repo.git", "https://github.com/user/repo"},
		{"git@gitlab.com:group/project.git", "https://gitlab.com/group/project"},
		{"git@bitbucket.org:team/repo.git", "https://bitbucket.org/team/repo"},
		{"git@ssh.dev.azure.com:v3/organization/project/repo", "https://dev.azure.com/organization/project/_git/repo"}, // Azure Repos
		{"user@github.com:user/repo.git", "https://github.com/user/repo"},                                              // ssh with custom username
		{"ssh://git@myserver.com:2222/user/repo.git", "https://myserver.com/user/repo"},                                // with port
		{"git+ssh://git@myserver.com/user/repo.git", "https://myserver.com/user/repo"},                                 // with git+ssh
		{"git::ssh://git@myserver.com/user/repo.git", "https://myserver.com/user/repo"},                                // with git::ssh
		{"https://github.com/user/repo.git", "https://github.com/user/repo"},                                           // https
		{"https://user@myserver.com/user/repo.git", "https://myserver.com/user/repo"},                                  // https with username
		{"git::https://github.com/user/repo.git", "https://github.com/user/repo"},                                      // git::https
		{"git::ssh://git@myserver.com/user/repo.git?ref=a094835", "https://myserver.com/user/repo?ref=a094835"},        // git::ssh with ref
	}

	for _, test := range tests {
		output, err := normalizeModuleURL(test.input)
		if err != nil {
			if test.expected != "" {
				t.Errorf("normalizeModuleURL(%q) returned an error: %v", test.input, err)
			}
		} else {
			if !reflect.DeepEqual(output, test.expected) {
				t.Errorf("normalizeModuleURL(%q) = %q, want %q", test.input, output, test.expected)
			}
		}
	}
}
