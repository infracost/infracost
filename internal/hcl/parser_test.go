package hcl

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
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

	parsers, err := LoadParsers(filepath.Dir(path), []string{}, newDiscardLogger(), OptionStopOnHCLError())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parser := newParser(filepath.Dir(path), newDiscardLogger())
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

	parsers, err := LoadParsers(path, []string{}, newDiscardLogger(), OptionStopOnHCLError())
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

	parsers, err := LoadParsers(path, []string{}, newDiscardLogger(), OptionStopOnHCLError())
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
	dir, err := ioutil.TempDir(os.TempDir(), "infracost")
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
	dir, err := ioutil.TempDir(os.TempDir(), "infracost")
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
