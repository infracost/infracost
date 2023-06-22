package hcl

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/sync"
)

func Test_BasicParsing(t *testing.T) {
	path := createTestFile("test.tf", `

locals {
	proxy = var.cats_mother
}

variable "cats_mother" {
	default = "boots"
}

provider "cats" {

}

resource "cats_cat" "mittens" {
	name = "mittens"
	special = true
}

resource "cats_kitten" "the-great-destroyer" {
	name = "the great destroyer"
    parent = cats_cat.mittens.name
}

data "cats_cat" "the-cats-mother" {
	name = local.proxy
}


`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	// variable
	variables := blocks.OfType("variable")
	require.Len(t, variables, 1)
	assert.Equal(t, "variable", variables[0].Type())
	require.Len(t, variables[0].Labels(), 1)
	assert.Equal(t, "cats_mother", variables[0].TypeLabel())
	defaultVal := variables[0].GetAttribute("default")
	require.NotNil(t, defaultVal)
	assert.Equal(t, cty.String, defaultVal.Value().Type())
	assert.Equal(t, "boots", defaultVal.Value().AsString())

	// provider
	providerBlocks := blocks.OfType("provider")
	require.Len(t, providerBlocks, 1)
	assert.Equal(t, "provider", providerBlocks[0].Type())
	require.Len(t, providerBlocks[0].Labels(), 1)
	assert.Equal(t, "cats", providerBlocks[0].TypeLabel())

	// resources
	resourceBlocks := blocks.OfType("resource")

	sort.Slice(resourceBlocks, func(i, j int) bool {
		return resourceBlocks[i].TypeLabel() < resourceBlocks[j].TypeLabel()
	})

	require.Len(t, resourceBlocks, 2)
	require.Len(t, resourceBlocks[0].Labels(), 2)

	assert.Equal(t, "resource", resourceBlocks[0].Type())
	assert.Equal(t, "cats_cat", resourceBlocks[0].TypeLabel())
	assert.Equal(t, "mittens", resourceBlocks[0].NameLabel())

	assert.Equal(t, "mittens", resourceBlocks[0].GetAttribute("name").Value().AsString())
	assert.True(t, resourceBlocks[0].GetAttribute("special").Value().True())

	assert.Equal(t, "resource", resourceBlocks[1].Type())
	assert.Equal(t, "cats_kitten", resourceBlocks[1].TypeLabel())
	assert.Equal(t, "the great destroyer", resourceBlocks[1].GetAttribute("name").Value().AsString())
	assert.Equal(t, "mittens", resourceBlocks[1].GetAttribute("parent").Value().AsString())

	// data
	dataBlocks := blocks.OfType("data")
	require.Len(t, dataBlocks, 1)
	require.Len(t, dataBlocks[0].Labels(), 2)

	assert.Equal(t, "data", dataBlocks[0].Type())
	assert.Equal(t, "cats_cat", dataBlocks[0].TypeLabel())
	assert.Equal(t, "the-cats-mother", dataBlocks[0].NameLabel())

	assert.Equal(t, "boots", dataBlocks[0].GetAttribute("name").Value().AsString())
}

func Test_UnsupportedAttributes(t *testing.T) {
	path := createTestFile("test.tf", `

resource "with_unsupported_attr" "test" {
	name = "mittens"
	special = true
	my_number = 4
}

resource "with_unsupported_attr" "test2" {
	name = "mittens"
	special = true
}

locals {
	value = with_unsupported_attr.test.does_not_exist
	value_nested = with_unsupported_attr.test.a_block.attr
	names = ["test1", "test3"]
}

output "exp" {
  value = {
    for name in local.names :
      name => with_unsupported_attr.test.does_not_exist
  }
}

output "loadbalancer"  {
    value = {
		"${var.env_dnsname}:${var.app_dnsname}" : with_unsupported_attr.test.ip_address
    }
}
`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	label := blocks.Matching(BlockMatcher{Type: "locals"})
	require.NotNil(t, label)
	mockedVal := label.GetAttribute("value").Value()
	require.Equal(t, cty.String, mockedVal.Type())
	assert.Equal(t, "value-mock", mockedVal.AsString())

	mockedVarObj := label.GetAttribute("value_nested").Value()
	require.Equal(t, cty.String, mockedVarObj.Type())
	assert.Equal(t, "value_nested-mock", mockedVarObj.AsString())

	output := blocks.Matching(BlockMatcher{Label: "exp", Type: "output"})
	require.NotNil(t, output)
	mockedObj := output.GetAttribute("value").Value()
	require.True(t, mockedObj.Type().IsObjectType())

	vm := mockedObj.AsValueMap()
	assert.Len(t, vm, 2)
	assert.Equal(t, mockedVal, vm["test1"])
	assert.Equal(t, mockedVal, vm["test3"])

	output2 := blocks.Matching(BlockMatcher{Label: "loadbalancer", Type: "output"})
	objectWithKeys := output2.GetAttribute("value").Value()
	require.True(t, objectWithKeys.Type().IsObjectType())
	keys := []string{}
	for k, v := range objectWithKeys.AsValueMap() {
		keys = append(keys, k)
		pieces := strings.Split(k, ":")
		require.Len(t, pieces, 2)

		assert.Equal(t, "value-mock", pieces[0])
		assert.Equal(t, "value-mock", pieces[1])
		assert.Equal(t, "value-mock", v.AsString())
	}

	require.Len(t, keys, 1)

}

func Test_UnsupportedAttributesList(t *testing.T) {
	path := createTestFile("test.tf", `

resource "with_unsupported_attr" "test" {
	name = "mittens"
	special = true
	my_number = 4
}

locals {
	names = "astring"
}

output "exp" {
  value = {
    for name in local.names :
      name => with_unsupported_attr.test.does_not_exist
  }
}

output "exp2" {
  value = {
    for name in local.nothing :
      name => with_unsupported_attr.test.does_not_exist
  }
}
`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "exp", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	mockedObj := attr.Value()
	require.True(t, mockedObj.Type().IsObjectType())
	asMap := mockedObj.AsValueMap()
	assert.Len(t, asMap, 1)
	assert.Equal(t, "value-mock", asMap["astring"].AsString())

	output = blocks.Matching(BlockMatcher{Label: "exp2", Type: "output"})
	require.NotNil(t, output)
	attr = output.GetAttribute("value")
	mockedObj = attr.Value()
	require.True(t, mockedObj.Type().IsObjectType())
	asMap = mockedObj.AsValueMap()
	assert.Len(t, asMap, 1)
	for k, v := range asMap {
		assert.Equal(t, "value-mock", k)
		assert.Equal(t, "value-mock", v.AsString())
	}
}

func Test_UnsupportedAttributesInForeachIf(t *testing.T) {
	path := createTestFile("test.tf", `

data "google_compute_instance" "default" {
  count = 1
  name = "primary-application-server"
  zone = "us-central1-a"
}

output "instances" {
  value = {
    for instance in data.google_compute_instance.default :
    instance.name => instance.network_interface[0].network_ip
    if instance.network_interface[0].network_ip != null
  }
}
`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "instances", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	mockedObj := attr.Value()
	require.True(t, mockedObj.Type().IsObjectType())

	asMap := mockedObj.AsValueMap()
	assert.Len(t, asMap, 1)
	assert.Equal(t, "value-mock", asMap["primary-application-server"].AsString())
}

func Test_UnsupportedAttributesSplatOperator(t *testing.T) {
	path := createTestFile("test.tf", `

variable "enabled" {
	default = true
}

resource "with_unsupported_attr" "test" {
	count = 4

	name = "mittens"
	special = true
	my_number = 4
}

resource "other_resource" "test" {
	task_definition = "${join("|", with_unsupported_attr.test.*.family)}:${join("|", with_unsupported_attr.test.*.revision)}"
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "other_resource.test", Type: "resource"})
	require.NotNil(t, output)
	attr := output.GetAttribute("task_definition")
	mockedVal := attr.Value()
	str := mockedVal.AsString()
	assert.Contains(t, str, ":")
	pieces := strings.Split(strings.ReplaceAll(str, ":", "|"), "|")
	assert.Len(t, pieces, 8)

	for _, piece := range pieces {
		assert.Equal(t, "task_definition-mock", piece)
	}
}

func Test_UnsupportedSplat(t *testing.T) {
	path := createTestFile("test.tf", `

data "mydata" "default" {
  zone = "us-central1-a"
}

resource "aws_subnet" "private" {
	count = data.mydata.default.enabled != null ? 2 : 0
	cidr_block = "10.0.1.0/24"
}

output "private_subnets_cidr_blocks" {
  description = "List of cidr_blocks of private subnets"
  value       = aws_subnet.private.*.cidr_block
}

output "attr_not_exists" {
  description = "List of cidr_blocks of private subnets"
  value       = aws_subnet.private.*.test
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "private_subnets_cidr_blocks", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	mockedVal := attr.Value()
	require.True(t, mockedVal.Type().IsTupleType())
	list := mockedVal.AsValueSlice()
	assert.Len(t, list, 2)
	for _, value := range list {
		assert.Equal(t, "10.0.1.0/24", value.AsString())
	}
}

func Test_UnsupportedAttributesIndex(t *testing.T) {
	path := createTestFile("test.tf", `

variable "enabled" {
	default = true
}

resource "with_unsupported_attr" "test" {
	count = 4

	name = "mittens"
	special = true
	my_number = 4
}

resource "other_resource" "test" {
	task_definition = "${join("", [with_unsupported_attr.test[2].family])}:${with_unsupported_attr.test[1].family}"
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "other_resource.test", Type: "resource"})
	require.NotNil(t, output)
	attr := output.GetAttribute("task_definition")
	mockedVal := attr.Value()
	for _, v := range strings.Split(mockedVal.AsString(), ":") {
		assert.Equal(t, "task_definition-mock", v)
	}
}

func Test_UnsupportedAttributesMap(t *testing.T) {
	path := createTestFile("test.tf", `

variable "service_connections" {
  type = map(object({
    name              = string
    subscription_id   = string
    subscription_name = string
  }))
  nullable    = false
  default = {
    dev = {
		name = "dev"
		subscription_id = "test"
		subscription_name = "testname"
	}
    prod = {
		name = "prod"
		subscription_id = "test2"
		subscription_name = "test2name"
	}
  }
}

data "azuread_service_principal" "serviceendpoints" {
  for_each = var.service_connections

  display_name = "display-${each.value.subscription_id}"
}

output "serviceendpoint_principals" {
  value       = { for k, v in var.service_connections : k => data.azuread_service_principal.serviceendpoints[k].object_id }
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "azuread_service_principal", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	mockedObj := attr.Value()
	require.True(t, mockedObj.Type().IsObjectType())
	asMap := mockedObj.AsValueMap()
	require.Len(t, asMap, 2)

	_, ok := asMap["dev"]
	assert.True(t, ok)
	_, ok = asMap["prod"]
	assert.True(t, ok)

	assert.Equal(t, "value-mock", asMap["dev"].AsString())
	assert.Equal(t, "value-mock", asMap["prod"].AsString())
}

func Test_UnsupportedAttributesLocalIndex(t *testing.T) {
	path := createTestFile("test.tf", `
variable "test" {}

locals {
  val = format("%s", var.test.id[0])
}

output "val" {
  value = local.val
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "val", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	value := attr.Value()
	require.True(t, value.Type().IsPrimitiveType(), "value is not primitive type but %s", value.Type().GoString())
	assert.Equal(t, "val-mock", value.AsString())
}

func Test_SetsHasChangesOnMod(t *testing.T) {
	path := createTestFile("test.tf", `variable "foo" {}`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path), HasChanges: true}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	assert.True(t, module.HasChanges)
}

func Test_UnsupportedAttributesMapIndex(t *testing.T) {
	path := createTestFile("test.tf", `
variable "test" {}

locals {
  val = format("%s", var.test.id["foo"])
}

output "val" {
  value = local.val
}

`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	output := blocks.Matching(BlockMatcher{Label: "val", Type: "output"})
	require.NotNil(t, output)
	attr := output.GetAttribute("value")
	value := attr.Value()
	require.True(t, value.Type().IsPrimitiveType(), "value is not primitive type but %s", value.Type().GoString())
	assert.Equal(t, "val-mock", value.AsString())
}

func Test_Modules(t *testing.T) {

	path := createTestFileWithModule(`
module "my-mod" {
	source = "../module"
	input = "ok"
}

output "result" {
	value = module.my-mod.mod_result
}
`,
		`
variable "input" {
	default = "?"
}

output "mod_result" {
	value = var.input
}
`,
		"module",
	)

	logger := newDiscardLogger()
	dir := filepath.Dir(path)
	loader := modules.NewModuleLoader(dir, nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(path, loader, nil, logger)
	require.NoError(t, err)
	rootModule, err := parsers[0].ParseDirectory()
	require.NoError(t, err)
	childModule := rootModule.Modules[0]

	moduleBlocks := rootModule.Blocks.OfType("module")
	require.Len(t, moduleBlocks, 1)

	assert.Equal(t, "module", moduleBlocks[0].Type())
	assert.Equal(t, "module.my-mod", moduleBlocks[0].FullName())
	inputAttr := moduleBlocks[0].GetAttribute("input")
	require.NotNil(t, inputAttr)
	require.Equal(t, cty.String, inputAttr.Value().Type())
	assert.Equal(t, "ok", inputAttr.Value().AsString())

	rootOutputs := rootModule.Blocks.OfType("output")
	require.Len(t, rootOutputs, 1)
	assert.Equal(t, "output.result", rootOutputs[0].FullName())
	valAttr := rootOutputs[0].GetAttribute("value")
	require.NotNil(t, valAttr)
	require.Equal(t, cty.String, valAttr.Value().Type())
	assert.Equal(t, "ok", valAttr.Value().AsString())

	childOutputs := childModule.Blocks.OfType("output")
	require.Len(t, childOutputs, 1)
	assert.Equal(t, "module.my-mod.output.mod_result", childOutputs[0].FullName())
	childValAttr := childOutputs[0].GetAttribute("value")
	require.NotNil(t, childValAttr)
	require.Equal(t, cty.String, childValAttr.Value().Type())
	assert.Equal(t, "ok", childValAttr.Value().AsString())

}

func Test_NestedParentModule(t *testing.T) {

	path := createTestFileWithModule(`
module "my-mod" {
	source = "../."
	input = "ok"
}

output "result" {
	value = module.my-mod.mod_result
}
`,
		`
variable "input" {
	default = "?"
}

output "mod_result" {
	value = var.input
}
`,
		"",
	)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(path, loader, nil, logger)
	require.NoError(t, err)
	rootModule, err := parsers[0].ParseDirectory()
	require.NoError(t, err)
	childModule := rootModule.Modules[0]

	moduleBlocks := rootModule.Blocks.OfType("module")
	require.Len(t, moduleBlocks, 1)

	assert.Equal(t, "module", moduleBlocks[0].Type())
	assert.Equal(t, "module.my-mod", moduleBlocks[0].FullName())
	inputAttr := moduleBlocks[0].GetAttribute("input")
	require.NotNil(t, inputAttr)
	require.Equal(t, cty.String, inputAttr.Value().Type())
	assert.Equal(t, "ok", inputAttr.Value().AsString())

	rootOutputs := rootModule.Blocks.OfType("output")
	require.Len(t, rootOutputs, 1)
	assert.Equal(t, "output.result", rootOutputs[0].FullName())
	valAttr := rootOutputs[0].GetAttribute("value")
	require.NotNil(t, valAttr)
	require.Equal(t, cty.String, valAttr.Value().Type())
	assert.Equal(t, "ok", valAttr.Value().AsString())

	childOutputs := childModule.Blocks.OfType("output")
	require.Len(t, childOutputs, 1)
	assert.Equal(t, "module.my-mod.output.mod_result", childOutputs[0].FullName())
	childValAttr := childOutputs[0].GetAttribute("value")
	require.NotNil(t, childValAttr)
	require.Equal(t, cty.String, childValAttr.Value().Type())
	assert.Equal(t, "ok", childValAttr.Value().AsString())
}

func Test_LocalsObjectType(t *testing.T) {
	path := createTestFile("test.tf", `
data "aws_ami" "my_ami" {
  count = 1
}

locals {
  defaults = {
    platform = "linux"
    ami = data.aws_ami.my_ami.*.bad[0]
  }
}

resource "aws_instance" "my_instance" {
	platform = local.defaults.platform
  ami = local.defaults.ami
}
`)

	logger := newDiscardLogger()
	parser := newParser(RootPath{Path: filepath.Dir(path)}, modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{}), logger)
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	resource := blocks.Matching(BlockMatcher{Label: "test", Type: "resource"})
	require.NotNil(t, resource)

	platformAttr := resource.GetAttribute("platform")
	require.NotNil(t, platformAttr)
	assert.Equal(t, "linux", platformAttr.Value().AsString())

	amiAttr := resource.GetAttribute("ami")
	require.NotNil(t, amiAttr)
	assert.Equal(t, "defaults-mock", amiAttr.Value().AsString())
}

func TestOptionWithRawCtyInput(t *testing.T) {
	type args struct {
		input cty.Value
	}
	tests := []struct {
		name string
		args args
		want map[string]cty.Value
	}{
		{
			name: "test panic returns empty Option",
			args: args{
				input: cty.NilVal,
			},
			want: map[string]cty.Value{},
		},
		{
			name: "sets input vars from cty object",
			args: args{
				input: cty.ObjectVal(map[string]cty.Value{
					"test": cty.StringVal("val"),
				}),
			},
			want: map[string]cty.Value{
				"test": cty.StringVal("val"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Parser{inputVars: map[string]cty.Value{}}
			option := OptionWithRawCtyInput(tt.args.input)
			option(&p)

			assert.Equalf(t, tt.want, p.inputVars, "OptionWithRawCtyInput(%v)", tt.args.input)
		})
	}
}

func createTestFile(filename, contents string) string {
	dir, err := os.MkdirTemp(os.TempDir(), "infracost")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(contents), os.ModePerm); err != nil {
		panic(err)
	}
	return path
}

func createTestFileWithModule(contents string, moduleContents string, moduleName string) string {
	dir, err := os.MkdirTemp(os.TempDir(), "infracost")
	if err != nil {
		panic(err)
	}

	rootPath := filepath.Join(dir, "main")
	modulePath := dir
	if len(moduleName) > 0 {
		modulePath = filepath.Join(modulePath, moduleName)
	}

	if err := os.Mkdir(rootPath, 0755); err != nil {
		panic(err)
	}

	if modulePath != dir {
		if err := os.Mkdir(modulePath, 0755); err != nil {
			panic(err)
		}
	}

	if err := os.WriteFile(filepath.Join(rootPath, "main.tf"), []byte(contents), os.ModePerm); err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(modulePath, "main.tf"), []byte(moduleContents), os.ModePerm); err != nil {
		panic(err)
	}

	return rootPath
}

func newDiscardLogger() *logrus.Entry {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l.WithFields(logrus.Fields{})
}

func Test_NestedForEach(t *testing.T) {
	path := createTestFile("test.tf", `

locals {
  az = {
    a = {
      az = "us-east-1a"
      bz = "w"
    }
    b = {
      az = "us-east-1b"
      bz = "z"
    }
  }
}

resource "test_resource" "test" {
  for_each = local.az
  availability_zone       = each.value.az
  another_attr = "attr-${each.value.bz}"
}

resource "test_resource" "static" {
  availability_zone       = "az-static"
  another_attr = "attr-static"
}

resource "test_resource_two" "test" {
  for_each        = test_resource.test
  inherited_id    = each.value.id
  inherited_attr  = each.value.another_attr
}
`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	resources := blocks.OfType("resource")
	labels := make([]string, len(resources))
	for i, resource := range resources {
		labels[i] = resource.Label()
	}
	assert.ElementsMatch(t, []string{
		`test_resource.test["a"]`,
		`test_resource.test["b"]`,
		`test_resource.static`,
		`test_resource_two.test["a"]`,
		`test_resource_two.test["b"]`,
	}, labels)
}

func Test_ModuleForEaches(t *testing.T) {

	path := createTestFileWithModule(`
locals {
  az = {
    a = {
      az = "us-east-1a"
      bz = "w"
    }
    b = {
      az = "us-east-1b"
      bz = "z"
    }
  }
}

module "test" {
    for_each = local.az
	source = "../."
	input = each.value.bz
}

resource "test_two" "test" {
  for_each        = module.test
  inherited_id    = each.value.id
  inherited_attr  = each.value.mod_result
}
`,
		`
variable "input" {
	default = "?"
}

output "mod_result" {
	value = var.input
}
`,
		"",
	)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(path, loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	resources := blocks.OfType("resource")

	var ar *Block
	var br *Block
	for _, resource := range resources {
		if resource.Label() == "test_two.test[\"a\"]" {
			ar = resource
		}

		if resource.Label() == "test_two.test[\"b\"]" {
			br = resource
		}
	}
	require.NotNil(t, ar, "test_two.test[\"a\"] was not found in module blocks")
	require.NotNil(t, br, "test_two.test[\"b\"] was not found in module blocks")

	s := ar.GetAttribute("inherited_attr").AsString()
	assert.Equal(t, "w", s)

	s = br.GetAttribute("inherited_attr").AsString()
	assert.Equal(t, "z", s)

	modules := blocks.OfType("module")
	modLabels := make([]string, len(modules))
	for i, module := range modules {
		modLabels[i] = "module." + module.Label()
	}

	assert.ElementsMatch(t, []string{
		`module.test["a"]`,
		`module.test["b"]`,
	}, modLabels)

	var a *Module
	for _, m := range module.Modules {
		if m.Name == `module.test["a"]` {
			a = m
		}
	}
	require.NotNil(t, a, "could not find module.test[a] in root module")
	out := a.Blocks.Outputs(false)
	v := out.AsValueMap()
	s = v["mod_result"].AsString()
	assert.Equal(t, "w", s)

	var b *Module
	for _, m := range module.Modules {
		if m.Name == `module.test["b"]` {
			b = m
		}
	}
	require.NotNil(t, b, "could not find module.test[b] in root module")
	out = b.Blocks.Outputs(false)
	v = out.AsValueMap()
	s = v["mod_result"].AsString()
	assert.Equal(t, "z", s)
}

func Test_ModuleExpansionBehindComplexExpression(t *testing.T) {
	path := createTestFileWithModule(`
variable "var1" {
  default = {
	"foo" = {
		"initial_prop" = {
			"name" = "test"
		}
    }
	"bar" = {
		"initial_prop" = {
			"name" = "test2"
		}
    }
  }
}

variable "var2" {
  default = {
	"foo" = {
		"merged_prop" = {
			"another_name" = "test_again"
		}
    }
	"bar" = {
		"merged_prop" = {
			"another_name" = "test_again_2"
		}
    }
  }
}

locals {
  to_merge = { for to_merge_key, to_merge_value in var.var1 : to_merge_key => {
    merged       = module.test[local.index_prop].mod_result.out.obj
    }
  }

  merged = { for k, v in var.var1 : k => merge(v, local.to_merge[k]) }
}

locals {
  index_prop    =  "foo"
}

module "test" {
  for_each = var.var2

  source                      = "../."
  name                        = each.key
  obj                   	  = each.value.merged_prop
}

module "test2" {
	for_each = local.merged
  	source                      = "../."

    name                        = each.key
	obj                   	  	= each.value
}
`,
		`
variable "name" {
	default = "?"
}

variable "obj" {
	default = "?"
}

output "mod_result" {
	value = {
		"out" = {
			"name": var.name
			"obj": var.obj
		}
	}
}
`,
		"",
	)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(path, loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	var mod1 *Module
	var mod2 *Module
	for _, m := range module.Modules {
		if m.Name == `module.test["foo"]` {
			mod1 = m
		}

		if m.Name == `module.test2["foo"]` {
			mod2 = m
		}
	}

	assert.Len(t, module.Modules, 4)
	require.NotNil(t, mod1, "could not find module.test[foo] in root module")
	require.NotNil(t, mod2, "could not find module.test2[foo] in root module")
	out := mod1.Blocks.Outputs(false)
	v := out.AsValueMap()
	output := v["mod_result"]
	simple := ctyJson.SimpleJSONValue{Value: output}
	b, err := simple.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"out":{"name":"foo","obj":{"another_name":"test_again"}}}`, string(b))

	out = mod2.Blocks.Outputs(false)
	v = out.AsValueMap()
	output = v["mod_result"]
	simple = ctyJson.SimpleJSONValue{Value: output}
	b, err = simple.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{"out":{"name":"foo","obj":{"initial_prop":{"name":"test"},"merged":{"another_name":"test_again"}}}}`, string(b))
}

func Test_DynamicBlockWithMockedIndex(t *testing.T) {
	path := createTestFileWithModule(`
data "bad_state" "bad" {}

module "reload" {
  source         = "../."
  input = [
    {
      "ip"        = "10.0.0.0"
      "mock" = data.bad_state.bad.my_bad["10.0.0.0/24"].id
    },
    {
      "ip"        = "10.0.1.0"
      "mock" = data.bad_state.bad.my_bad["10.0.0.0/24"].id
    }
  ]
}

`,
		`
variable "input" {}

resource "dynamic" "resource" {
  dynamic "child_block" {
    for_each = {
    	for i in var.input : i.ip => i
    }

    content {
      bar = child_block.value.mock
      foo = child_block.value.ip
    }
  }
}
`,
		"",
	)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(path, loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	require.Len(t, module.Modules, 1)
	mod1 := module.Modules[0]
	resource := mod1.Blocks.Matching(BlockMatcher{Label: "dynamic.resource"})
	children := resource.GetChildBlocks("child_block")

	values := `[`
	for _, value := range children {
		b := valueToBytes(t, value.Values())
		values += string(b) + ","
	}
	values = strings.TrimSuffix(values, ",") + "]"

	assert.JSONEq(
		t,
		values,
		`[
			{"foo":"10.0.0.0","bar":"input-mock"},
			{"foo":"10.0.1.0","bar":"input-mock"}
		]`,
	)

}

func Test_ForEachReferencesAnotherForEachDependentAttribute(t *testing.T) {
	path := createTestFile("test.tf", `
locals {
  os_types = ["Windows"]
  skus     = ["EP1"]

  permutations = distinct(flatten([
	  for os_type in local.os_types : [
		  for sku in local.skus :{
			sku     = sku
			os_type = os_type
		  }
	  ]
  ]))
}

resource "azurerm_service_plan" "plan" {
  for_each = {for entry in local.permutations : "${entry.os_type}.${entry.sku}" => entry}

  name                = "plan-${each.value.os_type}-${each.value.sku}"
}

resource "azurerm_linux_function_app" "function" {
  for_each = {for entry in azurerm_service_plan.plan : "${entry.name}" => entry}

  name                       = each.value.name
}
`,
	)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	resource := module.Blocks.Matching(BlockMatcher{Label: `azurerm_linux_function_app.function["plan-Windows-EP1"]`})
	name := resource.GetAttribute("name").AsString()
	assert.Equal(t, "plan-Windows-EP1", name)
}

func valueToBytes(t *testing.T, v cty.Value) []byte {
	t.Helper()

	simple := ctyJson.SimpleJSONValue{Value: v}
	b, err := simple.MarshalJSON()
	require.NoError(t, err)

	return b
}

func assertBlockEqualsJSON(t *testing.T, expected string, actual cty.Value, remove ...string) {
	t.Helper()

	vals := actual.AsValueMap()
	for _, s := range remove {
		delete(vals, s)
	}

	b := valueToBytes(t, cty.ObjectVal(vals))
	assert.JSONEq(t, expected, string(b))
}

func Test_CountOutOfOrder(t *testing.T) {
	path := createTestFile("test.tf", `

resource "test_resource" "first" {
	count = length(test_resource.second)
}

resource "test_resource" "second" {
  count = 2
}
`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	resources := blocks.OfType("resource")
	labels := make([]string, len(resources))
	for i, resource := range resources {
		labels[i] = resource.Label()
	}
	assert.ElementsMatch(t, []string{
		`test_resource.first[0]`,
		`test_resource.first[1]`,
		`test_resource.second[0]`,
		`test_resource.second[1]`,
	}, labels)
}

func Test_ProvideMockZonesForGCPDataBlock(t *testing.T) {
	path := createTestFile("test.tf", `
provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "europe-west2"
}

provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-east1"
  alias       = "east1"
}

variable "regions" {
  default = ["us-east4", "me-central1"]
}

data "google_compute_zones" "variable" {
  for_each = toset(var.regions)
  region   = each.value
}

data "google_compute_zones" "eu" {}

data "google_compute_zones" "us" {
	provider = google.east1
}
`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	eu := blocks.Matching(BlockMatcher{Label: "google_compute_zones.eu", Type: "data"})
	b := valueToBytes(t, eu.Values())
	assert.JSONEq(t, `{"names":["europe-west2-a","europe-west2-b","europe-west2-c"]}`, string(b))

	us := blocks.Matching(BlockMatcher{Label: "google_compute_zones.us", Type: "data"})
	b = valueToBytes(t, us.Values())
	assert.JSONEq(t, `{"names":["us-east1-b","us-east1-c","us-east1-d"]}`, string(b))

	us4 := blocks.Matching(BlockMatcher{Label: `google_compute_zones.variable["us-east4"]`, Type: "data"})
	b = valueToBytes(t, us4.Values())
	assert.JSONEq(t, `{"names":["us-east4-a","us-east4-b","us-east4-c"]}`, string(b))

	me := blocks.Matching(BlockMatcher{Label: `google_compute_zones.variable["me-central1"]`, Type: "data"})
	b = valueToBytes(t, me.Values())
	assert.JSONEq(t, `{"names":["me-central1-a","me-central1-b","me-central1-c"]}`, string(b))
}

func Test_ProvideMockZonesForAWSDataBlock(t *testing.T) {
	path := createTestFile("test.tf", `
provider "aws" {
  region                      = "us-east-1"
}

provider "aws" {
  alias 					  = "eu"
  region                      = "eu-west-2"
}

data "aws_availability_zones" "us" {}

data "aws_availability_zones" "eu" {
	provider = aws.eu
}
`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	eu := blocks.Matching(BlockMatcher{Label: "aws_availability_zones.eu", Type: "data"})
	b := valueToBytes(t, eu.Values())
	assert.JSONEq(t, `{"group_names":["eu-west-2","eu-west-2","eu-west-2","eu-west-2-wl1","eu-west-2-wl1"],"id":"eu-west-2","names":["eu-west-2a","eu-west-2b","eu-west-2c","eu-west-2-wl1-lon-wlz-1","eu-west-2-wl1-man-wlz-1"],"zone_ids":["euw2-az2","euw2-az3","euw2-az1","euw2-wl1-lon-wlz1","euw2-wl1-man-wlz1"]}`, string(b))

	us := blocks.Matching(BlockMatcher{Label: "aws_availability_zones.us", Type: "data"})
	b = valueToBytes(t, us.Values())
	assert.JSONEq(t, `{"group_names":["us-east-1","us-east-1","us-east-1","us-east-1","us-east-1","us-east-1","us-east-1-atl-1","us-east-1-bos-1","us-east-1-bue-1","us-east-1-chi-1","us-east-1-dfw-1","us-east-1-iah-1","us-east-1-lim-1","us-east-1-mci-1","us-east-1-mia-1","us-east-1-msp-1","us-east-1-nyc-1","us-east-1-phl-1","us-east-1-qro-1","us-east-1-scl-1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1","us-east-1-wl1"],"id":"us-east-1","names":["us-east-1a","us-east-1b","us-east-1c","us-east-1d","us-east-1e","us-east-1f","us-east-1-atl-1a","us-east-1-bos-1a","us-east-1-bue-1a","us-east-1-chi-1a","us-east-1-dfw-1a","us-east-1-iah-1a","us-east-1-lim-1a","us-east-1-mci-1a","us-east-1-mia-1a","us-east-1-msp-1a","us-east-1-nyc-1a","us-east-1-phl-1a","us-east-1-qro-1a","us-east-1-scl-1a","us-east-1-wl1-atl-wlz-1","us-east-1-wl1-bna-wlz-1","us-east-1-wl1-bos-wlz-1","us-east-1-wl1-chi-wlz-1","us-east-1-wl1-clt-wlz-1","us-east-1-wl1-dfw-wlz-1","us-east-1-wl1-dtw-wlz-1","us-east-1-wl1-iah-wlz-1","us-east-1-wl1-mia-wlz-1","us-east-1-wl1-msp-wlz-1","us-east-1-wl1-nyc-wlz-1","us-east-1-wl1-tpa-wlz-1","us-east-1-wl1-was-wlz-1"],"zone_ids":["use1-az6","use1-az1","use1-az2","use1-az4","use1-az3","use1-az5","use1-atl1-az1","use1-bos1-az1","use1-bue1-az1","use1-chi1-az1","use1-dfw1-az1","use1-iah1-az1","use1-lim1-az1","use1-mci1-az1","use1-mia1-az1","use1-msp1-az1","use1-nyc1-az1","use1-phl1-az1","use1-qro1-az1","use1-scl1-az1","use1-wl1-atl-wlz1","use1-wl1-bna-wlz1","use1-wl1-bos-wlz1","use1-wl1-chi-wlz1","use1-wl1-clt-wlz1","use1-wl1-dfw-wlz1","use1-wl1-dtw-wlz1","use1-wl1-iah-wlz1","use1-wl1-mia-wlz1","use1-wl1-msp-wlz1","use1-wl1-nyc-wlz1","use1-wl1-tpa-wlz1","use1-wl1-was-wlz1"]}`, string(b))
}

func Test_RandomShuffleSetsResult(t *testing.T) {
	path := createTestFile("test.tf", `
resource "random_shuffle" "one" {
  input        = ["a", "b", "c"]
  result_count = 1
}

resource "random_shuffle" "two" {
  input        = ["a", "b", "c"]
  result_count = 2
}

resource "random_shuffle" "nil" {
  input        = ["a", "b", "c"]
}

resource "random_shuffle" "large" {
  input        = ["a", "b", "c"]
  result_count = 5
}

resource "random_shuffle" "bad" {
  input        = 3
}
`)

	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(filepath.Dir(path), nil, config.TerraformSourceMap{}, logger, &sync.KeyMutex{})
	parsers, err := LoadParsers(filepath.Dir(path), loader, nil, logger)
	require.NoError(t, err)
	module, err := parsers[0].ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks
	assertBlockEqualsJSON(
		t,
		`{"input":["a","b","c"],"result":["a"],"result_count":1}`,
		blocks.Matching(BlockMatcher{Label: "random_shuffle.one", Type: "resource"}).Values(),
		"id", "arn", "self_link", "name",
	)
	assertBlockEqualsJSON(
		t,
		`{"input":["a","b","c"],"result":["a", "b"],"result_count":2}`,
		blocks.Matching(BlockMatcher{Label: "random_shuffle.two", Type: "resource"}).Values(),
		"id", "arn", "self_link", "name",
	)
	assertBlockEqualsJSON(
		t,
		`{"input":["a","b","c"],"result":["a", "b", "c"]}`,
		blocks.Matching(BlockMatcher{Label: "random_shuffle.nil", Type: "resource"}).Values(),
		"id", "arn", "self_link", "name",
	)
	assertBlockEqualsJSON(
		t,
		`{"input":["a","b","c"],"result":["a", "b", "c"],"result_count":5}`,
		blocks.Matching(BlockMatcher{Label: "random_shuffle.large", Type: "resource"}).Values(),
		"id", "arn", "self_link", "name",
	)
	assertBlockEqualsJSON(
		t,
		`{"input":3}`,
		blocks.Matching(BlockMatcher{Label: "random_shuffle.bad", Type: "resource"}).Values(),
		"id", "arn", "self_link", "name",
	)
}
