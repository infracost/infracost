package hcl

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	yaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/funcs"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/ui"
)

var (
	errorNoVarValue = errors.New("no value found")
	modReplace      = regexp.MustCompile(`^module\.`)
	// Use a separate regex for nested modules since we want to replace this with '.'
	nestedModReplace    = regexp.MustCompile(`\.module\.`)
	modArrayPartReplace = regexp.MustCompile(`\[[^[]*\]`)
	validBlocksToExpand = map[string]struct{}{
		"resource": {},
		"module":   {},
		"dynamic":  {},
		"data":     {},
	}
)

const maxContextIterations = 32

// Evaluator provides a set of given Blocks with contextual information.
// Evaluator is an important step in retrieving Block values that can be used in the
// schema.Resource cost retrieval. Without Evaluator the Blocks provided only have shallow information
// within attributes and won't contain any evaluated variables or references.
type Evaluator struct {
	// ctx is the master Context for evaluating the current set of Blocks. This is extremely important
	// and gets slowly built up as the Evaluator runs across the list of Blocks.
	ctx *Context
	// inputVars are the given input variables for this Evaluator run. At the root module level these are variables
	// provided by the user as tfvars. Further down the config tree these input vars are module variables provided in
	// HCL attributes.
	inputVars map[string]cty.Value
	// moduleCalls are the modules that the list of Blocks call to. This is built at runtime.
	moduleCalls []*ModuleCall
	// moduleMetadata is a lookup map of where modules exist on the local filesystem. This is built as part of a
	// Terraform or Infracost init.
	moduleMetadata *modules.Manifest
	// visitedModules is a lookup map to hold information by the Evaluator of modules that it has already evaluated.
	visitedModules map[string]map[string]cty.Value
	// module defines the input and module path for the Evaluator. It is the root module of the config.
	module Module
	// workingDir is the current directory the evaluator is running within. This is used to set Context information on
	// child modules that the evaluator visits.
	workingDir string
	// workspace is the Terraform workspace that the Evaluator is running within.
	workspace string
	// blockBuilder handles generating blocks in the evaluation step.
	blockBuilder BlockBuilder
	newSpinner   ui.SpinnerFunc
	logger       *logrus.Entry
}

// NewEvaluator returns an Evaluator with Context initialised with top level variables.
// This Context is then passed to all Blocks as child Context so that variables built in Evaluation
// are propagated to the Block Attributes.
func NewEvaluator(
	module Module,
	workingDir string,
	inputVars map[string]cty.Value,
	moduleMetadata *modules.Manifest,
	visitedModules map[string]map[string]cty.Value,
	workspace string,
	blockBuilder BlockBuilder,
	spinFunc ui.SpinnerFunc,
	logger *logrus.Entry,
) *Evaluator {
	ctx := NewContext(&hcl.EvalContext{
		Functions: expFunctions(module.RootPath),
	}, nil, logger)

	if visitedModules == nil {
		visitedModules = make(map[string]map[string]cty.Value)
	}

	modulePath, err := filepath.Rel(module.RootPath, module.ModulePath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"rootPath":   module.RootPath,
			"modulePath": module.ModulePath,
		}).Debug("Could not calculate relative path.module")

		modulePath = module.ModulePath
	}

	// set the global evaluation parameters.
	ctx.SetByDot(cty.StringVal(workspace), "terraform.workspace")
	ctx.SetByDot(cty.StringVal(module.RootPath), "path.root")
	ctx.SetByDot(cty.StringVal(modulePath), "path.module")
	ctx.SetByDot(cty.StringVal(filepath.Join(workingDir, module.RootPath)), "path.cwd")

	for _, b := range module.Blocks {
		b.SetContext(ctx.NewChild())
	}

	moduleName := module.Name
	if moduleName == "" {
		moduleName = "root"
	}

	l := logger.WithFields(logrus.Fields{
		"module_name":   moduleName,
		"module_source": module.Source,
		"module_path":   module.ModulePath,
	})

	return &Evaluator{
		module:         module,
		ctx:            ctx,
		inputVars:      inputVars,
		moduleMetadata: moduleMetadata,
		visitedModules: visitedModules,
		workspace:      workspace,
		workingDir:     workingDir,
		blockBuilder:   blockBuilder,
		newSpinner:     spinFunc,
		logger:         l,
	}
}

// MissingVars returns a list of names of the variable blocks with missing input values.
func (e *Evaluator) MissingVars() []string {
	var missing []string

	blocks := e.module.Blocks.OfType("variable")
	for _, block := range blocks {
		_, v := e.evaluateVariable(block)
		if v == errorNoVarValue {
			missing = append(missing, fmt.Sprintf("'variable.%s'", block.Label()))
		}
	}

	return missing
}

// Run builds the Evaluator Context using all the provided Blocks. It will build up the Context to hold
// variable and reference information so that this can be used by Attribute evaluation. Run will also
// parse and build up and child modules that are referenced in the Blocks and runs child Evaluator on
// this Module.
func (e *Evaluator) Run() (*Module, error) {
	if e.newSpinner != nil {
		spin := e.newSpinner("Evaluating Terraform directory")
		defer spin.Success()
	}

	var lastContext hcl.EvalContext
	// first we need to evaluate the top level Context - so this can be passed to any child modules that are found.
	e.logger.Debug("evaluating top level context")
	e.evaluate(lastContext)

	// let's load the modules now we have our top level context.
	e.moduleCalls = e.loadModules()
	e.logger.Debug("evaluating context after loading modules")
	e.evaluate(lastContext)

	// expand out resources and modules via count and evaluate again so that we can include
	// any module outputs and or count references.
	e.module.Blocks = e.expandBlocks(e.module.Blocks)
	e.logger.Debug("evaluating context after expanding blocks")
	e.evaluate(lastContext)

	// returns all the evaluated Blocks under their given Module.
	return e.collectModules(), nil
}

func (e *Evaluator) collectModules() *Module {
	for _, definition := range e.moduleCalls {
		e.module.Modules = append(e.module.Modules, definition.Module)
	}

	return &e.module
}

// evaluate runs a context evaluation loop until the context values are unchanged. We run this in a loop
// because variables can change because of outputs from other blocks in the context. Once all outputs have
// been evaluated and the context variables should remain unchanged. In reality 90% of cases will require
// 2 loops, however other complex modules will take > 2.
func (e *Evaluator) evaluate(lastContext hcl.EvalContext) {
	for i := 0; i < maxContextIterations; i++ {
		e.evaluateStep(i)

		if reflect.DeepEqual(lastContext.Variables, e.ctx.Inner().Variables) {
			e.logger.Debug("evaluated outputs are the same a last context, exiting")
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

// evaluateStep gets the values for all the Block types in the current Module that affect Context.
// It then sets these values on the Context so that they can be used in Block Attribute evaluation.
func (e *Evaluator) evaluateStep(i int) {
	e.logger.Debugf("starting context evaluation iteration %d", i+1)

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

// evaluateModules loops over each of the moduleCalls in this Module and set a child Evaluator
// to run on the child Module Blocks. It passes the Evaluator the top level module Attributes as input variables.
func (e *Evaluator) evaluateModules() {
	for _, moduleCall := range e.moduleCalls {
		fullName := moduleCall.Definition.FullName()
		e.logger.Debugf("evaluating module call %s with source %s", fullName, moduleCall.Module.Source)
		vars := moduleCall.Definition.Values().AsValueMap()
		if oldVars, ok := e.visitedModules[fullName]; ok {
			if reflect.DeepEqual(vars, oldVars) {
				continue
			}

			e.logger.Debugf("module %s output vars have changed, evaluating again", fullName)
		}

		e.visitedModules[fullName] = vars

		moduleEvaluator := NewEvaluator(
			Module{
				Name:       fullName,
				Source:     moduleCall.Module.Source,
				Blocks:     moduleCall.Module.RawBlocks,
				RawBlocks:  moduleCall.Module.RawBlocks,
				RootPath:   e.module.RootPath,
				ModulePath: moduleCall.Path,
				Modules:    nil,
				Parent:     &e.module,
			},
			e.workingDir,
			vars,
			e.moduleMetadata,
			e.visitedModules,
			e.workspace,
			e.blockBuilder,
			nil,
			e.logger,
		)

		moduleCall.Module, _ = moduleEvaluator.Run()
		outputs := moduleEvaluator.exportOutputs()
		e.ctx.Set(outputs, "module", moduleCall.Name)
	}
}

// exportOutputs exports module outputs so that it can be used in Context evaluation.
func (e *Evaluator) exportOutputs() cty.Value {
	return e.module.Blocks.Outputs(false)
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
	for _, sub := range b.Children() {
		e.expandDynamicBlock(sub)
	}

	for _, sub := range b.Children().OfType("dynamic") {
		e.logger.Debugf("expanding block %s because a dynamic blocka was found %s", b.LocalName(), sub.LocalName())

		blockName := sub.TypeLabel()
		expanded := e.expandBlockForEaches([]*Block{sub})
		for _, ex := range expanded {
			if content := ex.GetChildBlock("content"); content != nil {
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
		if forEachAttr == nil || block.IsCountExpanded() || !shouldExpandBlock(block) {
			forEachFiltered = append(forEachFiltered, block)
			continue
		}

		e.logger.Debugf("expanding block %s because a for_each attribute was found", block.LocalName())

		if !forEachAttr.Value().IsNull() && forEachAttr.Value().IsKnown() && forEachAttr.IsIterable() {
			forEachAttr.Value().ForEachElement(func(key cty.Value, val cty.Value) bool {
				clone := e.blockBuilder.CloneBlock(block, key)

				ctx := clone.Context()

				e.copyVariables(block, clone)

				ctx.SetByDot(key, "each.key")
				ctx.SetByDot(val, "each.value")

				ctx.Set(key, block.TypeLabel(), "key")
				ctx.Set(val, block.TypeLabel(), "value")

				forEachFiltered = append(forEachFiltered, clone)

				return false
			})
		}
	}

	return forEachFiltered
}

func shouldExpandBlock(block *Block) bool {
	_, isValidType := validBlocksToExpand[block.Type()]
	return isValidType
}

func (e *Evaluator) expandBlockCounts(blocks Blocks) Blocks {
	var countFiltered Blocks
	for _, block := range blocks {
		countAttr := block.GetAttribute("count")
		if countAttr == nil || !block.ShouldExpand() {
			countFiltered = append(countFiltered, block)
			continue
		}

		count := 1
		value := countAttr.Value()
		if !value.IsNull() && value.IsKnown() {
			v := countAttr.AsInt()
			if v <= math.MaxInt32 {
				count = int(v)
			}
		}

		e.logger.Debugf("expanding block %s because a count attribute of value %d was found", block.LocalName(), count)

		vals := make([]cty.Value, count)
		for i := 0; i < count; i++ {
			c, _ := gocty.ToCtyValue(i, cty.Number)
			clone := e.blockBuilder.CloneBlock(block, c)

			countFiltered = append(countFiltered, clone)
			vals[i] = clone.Values()
		}

		e.ctx.SetByDot(cty.TupleVal(vals), block.Reference().String())
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
		e.logger.Debugf("error trying to copyVariable from the source of '%s.%s'", fromBase, fromRel)
		return
	}
	e.ctx.Root().Set(srcValue, fromBase, toRel)
}

func (e *Evaluator) evaluateVariable(b *Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.NilVal, fmt.Errorf("empty label - cannot resolve")
	}

	attributes := b.AttributesAsMap()
	if attributes == nil {
		return cty.NilVal, fmt.Errorf("cannot resolve variable with no attributes")
	}

	attribute := attributes["type"]
	if override, exists := e.inputVars[b.Label()]; exists {
		return convertType(override, attribute), nil
	}

	if def, exists := attributes["default"]; exists {
		return def.Value(), nil
	}

	return convertType(cty.NilVal, attribute), errorNoVarValue
}

func convertType(val cty.Value, attribute *Attribute) cty.Value {
	if attribute == nil {
		return val
	}

	var t string
	switch v := attribute.HCLAttr.Expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		t = v.Traversal.RootName()
	case *hclsyntax.LiteralValueExpr:
		t = attribute.AsString()
	}

	switch t {
	case "string":
		return valueToType(val, cty.String)
	case "number":
		return valueToType(val, cty.Number)
	case "bool":
		return valueToType(val, cty.Bool)
	}

	return val
}

func valueToType(val cty.Value, want cty.Type) cty.Value {
	if val == cty.NilVal {
		return val
	}

	newVal, err := convert.Convert(val, want)
	if err != nil {
		return val
	}

	return newVal
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

func (e *Evaluator) getValuesByBlockType(blockType string) cty.Value {
	blocksOfType := e.module.Blocks.OfType(blockType)
	values := make(map[string]cty.Value)

	for _, b := range blocksOfType {
		switch b.Type() {
		case "variable": // variables are special in that their value comes from the "default" attribute
			val, err := e.evaluateVariable(b)
			if err != nil {
				e.logger.WithError(err).Debugf("could not evaluate variable %s ignoring", b.FullName())
				continue
			}

			e.logger.Debugf("adding variable %s to the evaluation context", b.Label())
			values[b.Label()] = val
		case "output":
			val, err := e.evaluateOutput(b)
			if err != nil {
				e.logger.WithError(err).Debugf("could not evaluate output %s ignoring", b.FullName())
				continue
			}

			e.logger.Debugf("adding output %s to the evaluation context", b.Label())
			values[b.Label()] = val
		case "locals":
			for key, val := range b.Values().AsValueMap() {
				e.logger.Debugf("adding local %s to the evaluation context", key)

				values[key] = val
			}
		case "provider", "module":
			if b.Label() == "" {
				continue
			}

			e.logger.Debugf("adding %s %s to the evaluation context", b.Type(), b.Label())
			values[b.Label()] = b.Values()
		case "resource", "data":
			if len(b.Labels()) < 2 {
				continue
			}

			e.logger.Debugf("adding %s %s to the evaluation context", b.Type(), b.Label())
			values[b.Labels()[0]] = e.evaluateResource(b, values)
		}

	}

	return cty.ObjectVal(values)
}

func (e *Evaluator) evaluateResource(b *Block, values map[string]cty.Value) cty.Value {
	labels := b.Labels()

	blockMap, ok := values[labels[0]]
	if !ok {
		values[labels[0]] = cty.ObjectVal(make(map[string]cty.Value))
		blockMap = values[labels[0]]
	}

	valueMap := blockMap.AsValueMap()
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}

	if k := b.Key(); k != nil {
		e.logger.Debugf("expanding block %s to be available for for_each key %s", b.FullName(), *k)
		valueMap[stripCount(labels[1])] = expandedEachBlockToValue(b, valueMap)
		return cty.ObjectVal(valueMap)
	}

	if k := b.Index(); k != nil {
		e.logger.Debugf("expanding block %s to be available for index key %d", b.FullName(), *k)
		valueMap[stripCount(labels[1])] = expandCountBlockToValue(b, valueMap)
		return cty.ObjectVal(valueMap)
	}

	valueMap[b.Labels()[1]] = b.Values()
	return cty.ObjectVal(valueMap)
}

func expandCountBlockToValue(b *Block, existingValues map[string]cty.Value) cty.Value {
	k := b.Index()
	if k == nil {
		return cty.NilVal
	}

	vals := existingValues[stripCount(b.Labels()[1])]
	sourceTy := vals.Type()
	isList := sourceTy.IsTupleType() || sourceTy.IsListType() || sourceTy.IsSetType()

	var elements []cty.Value
	if isList {
		it := vals.ElementIterator()
		for it.Next() {
			_, v := it.Element()
			elements = append(elements, v)
		}
	}

	elements = append(elements, b.Values())
	return cty.TupleVal(elements)
}

func expandedEachBlockToValue(b *Block, existingValues map[string]cty.Value) cty.Value {
	k := b.Key()
	if k == nil {
		return cty.NilVal
	}

	ob := make(map[string]cty.Value)

	eachMap := existingValues[stripCount(b.Labels()[1])]
	if eachMap != cty.NilVal {
		for e, v := range eachMap.AsValueMap() {
			ob[e] = v
		}
	}

	ob[*k] = b.Values()
	return cty.ObjectVal(ob)
}

// loadModule takes in a module "x" {} block and loads resources etc. into e.moduleBlocks.
// Additionally, it returns variables to add to ["module.x.*"] variables
func (e *Evaluator) loadModule(b *Block) (*ModuleCall, error) {
	if b.Label() == "" {
		return nil, fmt.Errorf("module without label: %s", b.FullName())
	}

	var source string
	attrs := b.AttributesAsMap()
	for _, attr := range attrs {
		if attr.Name() == "source" {
			source = attr.AsString()
			break
		}
	}

	if source == "" {
		return nil, fmt.Errorf("could not read module source attribute at %s", b.FullName())
	}

	var modulePath string

	if e.moduleMetadata != nil {
		// if we have module metadata we can parse all the modules as they'll be cached locally!

		// Strip any "module." and "[*]" parts from the module name so it matches the manifest key format
		key := modReplace.ReplaceAllString(b.FullName(), "")
		key = nestedModReplace.ReplaceAllString(key, ".")
		key = modArrayPartReplace.ReplaceAllString(key, "")

		for _, module := range e.moduleMetadata.Modules {
			if module.Key == key {
				modulePath = filepath.Clean(filepath.Join(e.module.RootPath, module.Dir))
				break
			}
		}

		e.logger.Debugf("using path '%s' for module '%s' based on key '%s'", modulePath, b.FullName(), key)
	}

	if modulePath == "" {
		if !strings.HasPrefix(source, fmt.Sprintf(".%c", os.PathSeparator)) && !strings.HasPrefix(source, fmt.Sprintf("..%c", os.PathSeparator)) {
			reg := "registry.terraform.io/" + source
			return nil, fmt.Errorf("missing module with source '%s %s' -  try to 'terraform init' first", reg, source)
		}

		// combine the current calling module with relative source of the module
		modulePath = filepath.Join(e.module.ModulePath, source)
	}

	blocks, err := e.blockBuilder.BuildModuleBlocks(b, modulePath)
	if err != nil {
		return nil, err
	}
	e.logger.Debugf("loaded module '%s' (requested at %s)", modulePath, b.FullName())

	return &ModuleCall{
		Name:       b.Label(),
		Path:       modulePath,
		Definition: b,
		Module: &Module{
			Name:       b.TypeLabel(),
			Source:     source,
			Blocks:     blocks,
			RawBlocks:  blocks,
			RootPath:   e.module.RootPath,
			ModulePath: modulePath,
			Parent:     &e.module,
		},
	}, nil
}

// loadModules reads all module blocks and loads the underlying modules, adding blocks to moduleCalls.
func (e *Evaluator) loadModules() []*ModuleCall {
	e.logger.Debug("loading module calls")
	var moduleDefinitions []*ModuleCall

	// TODO: if a module uses a count that depends on a module output, then the block expansion might be incorrect.
	expanded := e.expandBlocks(e.module.Blocks.ModuleBlocks())

	for _, moduleBlock := range expanded {
		if moduleBlock.Label() == "" {
			continue
		}

		moduleCall, err := e.loadModule(moduleBlock)
		if err != nil {
			e.logger.WithError(err).Warnf("failed to load module %s ignoring", moduleBlock.LocalName())
			continue
		}

		moduleDefinitions = append(moduleDefinitions, moduleCall)
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
