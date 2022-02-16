package hcl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	log "github.com/sirupsen/logrus"
	yaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/funcs"
)

const maxContextIterations = 32

type Evaluator struct {
	ctx               *Context
	blocks            Blocks
	moduleDefinitions []*ModuleDefinition
	visitedModules    map[string]struct{}
	inputVars         map[string]cty.Value
	moduleMetadata    *ModulesMetadata
	projectRootPath   string // root of the current scan
	stopOnHCLError    bool
	modulePath        string
	workingDir        string
	workspace         string
}

func NewEvaluator(
	projectRootPath string,
	modulePath string,
	workingDir string,
	blocks Blocks,
	inputVars map[string]cty.Value,
	moduleMetadata *ModulesMetadata,
	visitedModules map[string]struct{},
	stopOnHCLError bool,
	workspace string,
) *Evaluator {
	ctx := NewContext(&hcl.EvalContext{
		Functions: expFunctions(modulePath),
	}, nil)

	if visitedModules == nil {
		visitedModules = make(map[string]struct{})
	}

	// set the global evaluation parameters.
	ctx.SetByDot(cty.StringVal(workspace), "terraform.workspace")
	ctx.SetByDot(cty.StringVal(projectRootPath), "path.root")
	ctx.SetByDot(cty.StringVal(modulePath), "path.module")
	ctx.SetByDot(cty.StringVal(workingDir), "path.cwd")

	for _, b := range blocks {
		b.OverrideContext(ctx.NewChild())
	}

	return &Evaluator{
		modulePath:      modulePath,
		projectRootPath: projectRootPath,
		workingDir:      workingDir,
		ctx:             ctx,
		blocks:          blocks,
		inputVars:       inputVars,
		moduleMetadata:  moduleMetadata,
		visitedModules:  visitedModules,
		stopOnHCLError:  stopOnHCLError,
		workspace:       workspace,
	}
}

func (e *Evaluator) EvaluateAll() ([]*Module, error) {
	var lastContext hcl.EvalContext
	e.evaluate(lastContext)
	e.moduleDefinitions = e.loadModules(true)

	// expand out resources and modules via count
	e.blocks = e.expandBlocks(e.blocks)
	e.evaluate(lastContext)

	return e.collectModules(), nil
}

func (e *Evaluator) evaluateStep(i int) {
	log.Debugf("Starting iteration %d of context evaluation...", i+1)

	e.ctx.Set(e.getValuesByBlockType("variable"), "var")
	e.ctx.Set(e.getValuesByBlockType("locals"), "local")
	e.ctx.Set(e.getValuesByBlockType("provider"), "provider")

	resources := e.getValuesByBlockType("resource")
	for key, resource := range resources.AsValueMap() {
		e.ctx.Set(resource, key)
	}

	e.ctx.Set(e.getValuesByBlockType("data"), "data")
	e.ctx.Set(e.getValuesByBlockType("output"), "output")

	e.evaluateModules()
}

func (e *Evaluator) evaluateModules() {
	for _, module := range e.moduleDefinitions {
		if _, ok := e.visitedModules[module.Definition.FullName()]; ok {
			continue
		}

		e.visitedModules[module.Definition.FullName()] = struct{}{}

		vars := module.Definition.Values().AsValueMap()
		moduleEvaluator := NewEvaluator(
			e.projectRootPath,
			module.Path,
			e.workingDir,
			module.Modules[0].Blocks,
			vars,
			e.moduleMetadata,
			e.visitedModules,
			e.stopOnHCLError,
			e.workspace,
		)
		module.Modules, _ = moduleEvaluator.EvaluateAll()

		e.ctx.Set(moduleEvaluator.exportOutputs(), "module", module.Name)
	}
}

// export module outputs to a parent
func (e *Evaluator) exportOutputs() cty.Value {
	data := make(map[string]cty.Value)

	for _, block := range e.blocks.OfType("output") {
		attr := block.GetAttribute("value")
		if attr == nil {
			continue
		}

		data[block.Label()] = attr.Value()
	}

	return cty.ObjectVal(data)
}

func (e *Evaluator) collectModules() []*Module {
	rootModule := &Module{Blocks: e.blocks, RootPath: e.projectRootPath, ModulePath: e.modulePath}
	modules := []*Module{rootModule}

	for _, definition := range e.moduleDefinitions {
		modules = append(modules, definition.Modules...)
	}

	return modules
}

func (e *Evaluator) evaluate(lastContext hcl.EvalContext) {
	// TODO: we need to work out why the evaluation step takes place in a loop. I have been unable to decipher
	// why we need to evaluate the context at least twice. No tests that we've seen can replicated the scenario
	// where lastContext.Variables change after the first `evaluateStep`. We should reach out to tfsec guys and
	// work out why this loop was put in there. If they can't come up with a good reason we can nuke it and
	// make the evaluate method much easier to understand.
	for i := 0; i < maxContextIterations; i++ {
		e.evaluateStep(i)

		if reflect.DeepEqual(lastContext.Variables, e.ctx.Inner().Variables) {
			break
		}

		if len(e.ctx.Inner().Variables) != len(lastContext.Variables) {
			lastContext.Variables = make(map[string]cty.Value, len(e.ctx.Inner().Variables))
		}

		for k, v := range e.ctx.Inner().Variables {
			lastContext.Variables[k] = v
		}
	}
}

func (e *Evaluator) expandBlocks(blocks Blocks) Blocks {
	return e.expandDynamicBlocks(e.expandBlockForEaches(e.expandBlockCounts(blocks))...)
}

func (e *Evaluator) expandDynamicBlocks(blocks ...*Block) Blocks {
	for _, b := range blocks {
		e.expandDynamicBlock(b)
	}
	return blocks
}

func (e *Evaluator) expandDynamicBlock(b *Block) {
	for _, sub := range b.AllBlocks() {
		e.expandDynamicBlock(sub)
	}

	for _, sub := range b.AllBlocks().OfType("dynamic") {
		blockName := sub.TypeLabel()
		expanded := e.expandBlockForEaches([]*Block{sub})
		for _, ex := range expanded {
			if content := ex.GetBlock("content"); content != nil {
				_ = e.expandDynamicBlocks(content)
				b.InjectBlock(content, blockName)
			}
		}
	}
}

func (e *Evaluator) expandBlockForEaches(blocks Blocks) Blocks {
	var forEachFiltered Blocks
	for _, block := range blocks {
		forEachAttr := block.GetAttribute("for_each")
		if forEachAttr == nil || block.IsCountExpanded() || (block.Type() != "resource" && block.Type() != "module" && block.Type() != "dynamic") {
			forEachFiltered = append(forEachFiltered, block)
			continue
		}

		if !forEachAttr.Value().IsNull() && forEachAttr.Value().IsKnown() && forEachAttr.IsIterable() {
			forEachAttr.Value().ForEachElement(func(key cty.Value, val cty.Value) bool {
				clone := block.Clone(key)

				ctx := clone.Context()

				e.copyVariables(block, clone)

				ctx.SetByDot(key, "each.key")
				ctx.SetByDot(val, "each.value")

				ctx.Set(key, block.TypeLabel(), "key")
				ctx.Set(val, block.TypeLabel(), "value")

				log.Debugf("Added %s from for_each", clone.Reference())
				forEachFiltered = append(forEachFiltered, clone)

				return false
			})
		}
	}

	return forEachFiltered
}

func (e *Evaluator) expandBlockCounts(blocks Blocks) Blocks {
	var countFiltered Blocks
	for _, block := range blocks {
		countAttr := block.GetAttribute("count")
		if countAttr == nil || block.IsCountExpanded() || (block.Type() != "resource" && block.Type() != "module") {
			countFiltered = append(countFiltered, block)
			continue
		}

		count := 1
		if !countAttr.Value().IsNull() && countAttr.Value().IsKnown() {
			if countAttr.Value().Type() == cty.Number {
				f, _ := countAttr.Value().AsBigFloat().Float64()
				count = int(f)
			}
		}

		for i := 0; i < count; i++ {
			c, _ := gocty.ToCtyValue(i, cty.Number)
			clone := block.Clone(c)
			block.TypeLabel()
			log.Debugf("Added %s from count var", clone.Reference())
			countFiltered = append(countFiltered, clone)
		}
	}

	return countFiltered
}

func (e *Evaluator) copyVariables(from, to *Block) {
	var fromBase string
	var fromRel string
	var toRel string

	switch from.Type() {
	case "resource":
		fromBase = from.TypeLabel()
		fromRel = from.NameLabel()
		toRel = to.NameLabel()
	case "module":
		fromBase = from.Type()
		fromRel = from.TypeLabel()
		toRel = to.TypeLabel()
	default:
		return
	}

	srcValue := e.ctx.Root().Get(fromBase, fromRel)
	if srcValue == cty.NilVal {
		log.Debugf("error trying to copyVariable from the source of '%s.%s'", fromBase, fromRel)
		return
	}
	e.ctx.Root().Set(srcValue, fromBase, toRel)
}

func (e *Evaluator) evaluateVariable(b *Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.NilVal, fmt.Errorf("empty label - cannot resolve")
	}

	attributes := b.Attributes()
	if attributes == nil {
		return cty.NilVal, fmt.Errorf("cannot resolve variable with no attributes")
	}

	if override, exists := e.inputVars[b.Label()]; exists {
		return override, nil
	}

	if def, exists := attributes["default"]; exists {
		return def.Value(), nil
	}

	return cty.NilVal, fmt.Errorf("no value found")
}

func (e *Evaluator) evaluateOutput(b *Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.NilVal, fmt.Errorf("empty label - cannot resolve")
	}

	attribute := b.GetAttribute("value")
	if attribute == nil {
		return cty.NilVal, fmt.Errorf("cannot resolve variable with no attributes")
	}
	return attribute.Value(), nil
}

// returns true if all evaluations were successful
func (e *Evaluator) getValuesByBlockType(blockType string) cty.Value {
	blocksOfType := e.blocks.OfType(blockType)
	values := make(map[string]cty.Value)

	for _, b := range blocksOfType {
		switch b.Type() {
		case "variable": // variables are special in that their value comes from the "default" attribute
			val, err := e.evaluateVariable(b)
			if err != nil {
				continue
			}
			values[b.Label()] = val
		case "output":
			val, err := e.evaluateOutput(b)
			if err != nil {
				continue
			}
			values[b.Label()] = val
		case "locals":
			for key, val := range b.Values().AsValueMap() {
				values[key] = val
			}
		case "provider", "module":
			if b.Label() == "" {
				continue
			}
			values[b.Label()] = b.Values()
		case "resource", "data":
			if len(b.Labels()) < 2 {
				continue
			}

			blockMap, ok := values[b.Label()]
			if !ok {
				values[b.Labels()[0]] = cty.ObjectVal(make(map[string]cty.Value))
				blockMap = values[b.Labels()[0]]
			}

			valueMap := blockMap.AsValueMap()
			if valueMap == nil {
				valueMap = make(map[string]cty.Value)
			}

			valueMap[b.Labels()[1]] = b.Values()
			values[b.Labels()[0]] = cty.ObjectVal(valueMap)
		}

	}

	return cty.ObjectVal(values)
}

// takes in a module "x" {} block and loads resources etc. into e.moduleBlocks - additionally returns variables to add to ["module.x.*"] variables
func (e *Evaluator) loadModule(b *Block, stopOnHCLError bool) (*ModuleDefinition, error) {
	if b.Label() == "" {
		return nil, fmt.Errorf("module without label: %s", b.FullName())
	}

	var source string
	attrs := b.Attributes()
	for _, attr := range attrs {
		if attr.Name() == "source" {
			sourceVal := attr.Value()
			if sourceVal.Type() == cty.String {
				source = sourceVal.AsString()
			}
		}
	}

	if source == "" {
		return nil, fmt.Errorf("could not read module source attribute at %s", b.FullName())
	}

	var modulePath string

	if e.moduleMetadata != nil {
		// if we have module metadata we can parse all the modules as they'll be cached locally!
		for _, module := range e.moduleMetadata.Modules {
			reg := "registry.terraform.io/" + source
			if module.Source == source || module.Source == reg {
				modulePath = filepath.Clean(filepath.Join(e.projectRootPath, module.Dir))
				break
			}
		}
	}

	if modulePath == "" {
		// if we have no metadata, we can only support modules available on the local filesystem
		// users wanting this feature should run a `terraform init` before running infracost to cache all modules locally
		if !strings.HasPrefix(source, fmt.Sprintf(".%c", os.PathSeparator)) && !strings.HasPrefix(source, fmt.Sprintf("..%c", os.PathSeparator)) {
			reg := "registry.terraform.io/" + source
			return nil, fmt.Errorf("missing module with source '%s %s' -  try to 'terraform init' first", reg, source)
		}

		// combine the current calling module with relative source of the module
		modulePath = filepath.Join(e.modulePath, source)
	}

	var blocks Blocks
	err := getModuleBlocks(b, modulePath, &blocks, stopOnHCLError)
	if err != nil {
		return nil, err
	}
	log.Debugf("Loaded module '%s' (requested at %s)", modulePath, b.FullName())

	return &ModuleDefinition{
		Name:       b.Label(),
		Path:       modulePath,
		Definition: b,
		Modules:    []*Module{{RootPath: e.projectRootPath, ModulePath: modulePath, Blocks: blocks}},
	}, nil
}

// LoadModules reads all module blocks and loads the underlying modules, adding blocks to e.moduleBlocks
func (e *Evaluator) loadModules(stopOnHCLError bool) []*ModuleDefinition {
	blocks := e.blocks
	var moduleDefinitions []*ModuleDefinition

	expanded := e.expandBlocks(blocks.OfType("module"))

	for _, moduleBlock := range expanded {
		if moduleBlock.Label() == "" {
			continue
		}

		moduleDefinition, err := e.loadModule(moduleBlock, stopOnHCLError)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: Failed to load module: %s\n", err)
			continue
		}

		moduleDefinitions = append(moduleDefinitions, moduleDefinition)
	}

	return moduleDefinitions
}

// expFunctions returns the set of functions that should be used to when evaluating
// expressions in the receiving scope.
func expFunctions(baseDir string) map[string]function.Function {
	return map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          funcs.AbsPathFunc,
		"basename":         funcs.BasenameFunc,
		"base64decode":     funcs.Base64DecodeFunc,
		"base64encode":     funcs.Base64EncodeFunc,
		"base64gzip":       funcs.Base64GzipFunc,
		"base64sha256":     funcs.Base64Sha256Func,
		"base64sha512":     funcs.Base64Sha512Func,
		"bcrypt":           funcs.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             stdlib.CeilFunc,
		"chomp":            stdlib.ChompFunc,
		"cidrhost":         funcs.CidrHostFunc,
		"cidrnetmask":      funcs.CidrNetmaskFunc,
		"cidrsubnet":       funcs.CidrSubnetFunc,
		"cidrsubnets":      funcs.CidrSubnetsFunc,
		"coalesce":         funcs.CoalesceFunc,
		"coalescelist":     stdlib.CoalesceListFunc,
		"compact":          stdlib.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         stdlib.ContainsFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          funcs.DirnameFunc,
		"distinct":         stdlib.DistinctFunc,
		"element":          stdlib.ElementFunc,
		"chunklist":        stdlib.ChunklistFunc,
		"file":             funcs.MakeFileFunc(baseDir, false),
		"fileexists":       funcs.MakeFileExistsFunc(baseDir),
		"fileset":          funcs.MakeFileSetFunc(baseDir),
		"filebase64":       funcs.MakeFileFunc(baseDir, true),
		"filebase64sha256": funcs.MakeFileBase64Sha256Func(baseDir),
		"filebase64sha512": funcs.MakeFileBase64Sha512Func(baseDir),
		"filemd5":          funcs.MakeFileMd5Func(baseDir),
		"filesha1":         funcs.MakeFileSha1Func(baseDir),
		"filesha256":       funcs.MakeFileSha256Func(baseDir),
		"filesha512":       funcs.MakeFileSha512Func(baseDir),
		"flatten":          stdlib.FlattenFunc,
		"floor":            stdlib.FloorFunc,
		"format":           stdlib.FormatFunc,
		"formatdate":       stdlib.FormatDateFunc,
		"formatlist":       stdlib.FormatListFunc,
		"indent":           stdlib.IndentFunc,
		"index":            funcs.IndexFunc, // stdlib.IndexFunc is not compatible
		"join":             stdlib.JoinFunc,
		"jsondecode":       stdlib.JSONDecodeFunc,
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             stdlib.KeysFunc,
		"length":           funcs.LengthFunc,
		"list":             funcs.ListFunc,
		"log":              stdlib.LogFunc,
		"lookup":           funcs.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"map":              funcs.MapFunc,
		"matchkeys":        funcs.MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              funcs.Md5Func,
		"merge":            stdlib.MergeFunc,
		"min":              stdlib.MinFunc,
		"parseint":         stdlib.ParseIntFunc,
		"pathexpand":       funcs.PathExpandFunc,
		"pow":              stdlib.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          funcs.ReplaceFunc,
		"reverse":          stdlib.ReverseListFunc,
		"rsadecrypt":       funcs.RsaDecryptFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       stdlib.SetProductFunc,
		"setsubtract":      stdlib.SetSubtractFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             funcs.Sha1Func,
		"sha256":           funcs.Sha256Func,
		"sha512":           funcs.Sha512Func,
		"signum":           stdlib.SignumFunc,
		"slice":            stdlib.SliceFunc,
		"sort":             stdlib.SortFunc,
		"split":            stdlib.SplitFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"timestamp":        funcs.TimestampFunc,
		"timeadd":          stdlib.TimeAddFunc,
		"title":            stdlib.TitleFunc,
		"tostring":         funcs.MakeToFunc(cty.String),
		"tonumber":         funcs.MakeToFunc(cty.Number),
		"tobool":           funcs.MakeToFunc(cty.Bool),
		"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"transpose":        funcs.TransposeFunc,
		"trim":             stdlib.TrimFunc,
		"trimprefix":       stdlib.TrimPrefixFunc,
		"trimspace":        stdlib.TrimSpaceFunc,
		"trimsuffix":       stdlib.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        funcs.URLEncodeFunc,
		"uuid":             funcs.UUIDFunc,
		"uuidv5":           funcs.UUIDV5Func,
		"values":           stdlib.ValuesFunc,
		"yamldecode":       yaml.YAMLDecodeFunc,
		"yamlencode":       yaml.YAMLEncodeFunc,
		"zipmap":           stdlib.ZipmapFunc,
	}

}
