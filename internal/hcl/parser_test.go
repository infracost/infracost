package hcl

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
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

	parsers, err := LoadParsers(filepath.Dir(path), []string{}, OptionStopOnHCLError())
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

	parser := newParser(filepath.Dir(path))
	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	blocks := module.Blocks

	label := blocks.Matching(BlockMatcher{Type: "locals"})
	require.NotNil(t, label)
	mockedVal := label.GetAttribute("value").Value()
	require.Equal(t, cty.String, mockedVal.Type())
	assertUUID(t, mockedVal.AsString())

	mockedVarObj := label.GetAttribute("value_nested").Value()
	require.Equal(t, cty.String, mockedVarObj.Type())
	assertUUID(t, mockedVarObj.AsString())

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

		assertUUID(t, pieces[0])
		assertUUID(t, pieces[1])
		assertUUID(t, v.AsString())
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

	parser := newParser(filepath.Dir(path))
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
	assertUUID(t, asMap["astring"].AsString())

	output = blocks.Matching(BlockMatcher{Label: "exp2", Type: "output"})
	require.NotNil(t, output)
	attr = output.GetAttribute("value")
	mockedObj = attr.Value()
	require.True(t, mockedObj.Type().IsObjectType())
	asMap = mockedObj.AsValueMap()
	assert.Len(t, asMap, 1)
	for k, v := range asMap {
		assertUUID(t, k)
		assertUUID(t, v.AsString())
	}
}

func assertUUID(t *testing.T, val string) {
	t.Helper()

	_, err := uuid.Parse(val)
	assert.NoError(t, err)
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

	parsers, err := LoadParsers(path, []string{}, OptionStopOnHCLError())
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

	parsers, err := LoadParsers(path, []string{}, OptionStopOnHCLError())
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
