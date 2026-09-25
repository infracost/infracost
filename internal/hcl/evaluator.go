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

	"github.com/hashicorp/go-cty-funcs/cidr"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	componentsFuncs "github.com/turbot/terraform-components/lang/funcs"

	"github.com/hashicorp/go-cty-funcs/crypto"
	"github.com/hashicorp/go-cty-funcs/encoding"
	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/rs/zerolog"
	yaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/funcs"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
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

	sensitiveRegxp = regexp.MustCompile(strings.Join([]string{
		"^image[_|-]",
		"[_|-]image$",
		"saml[_|-]role$",
		"secrets$",
		"aws[_|-]profile$",
		"key[_|-]id$",
		"secret[_|-]key$",
		"access[_|-]key$",
		"[_|-]token$",
		"^token[_|-]",
		"[_|-]secret$",
		"^secret[_|-]",
		"[_|-]password$",
		"^password[_|-]",
		"[_|-]username$",
		"^username[_|-]",
		"api[_|-]key",
		"expiration[_|-]date",
	}, "|"))
)

const maxContextIterations = 120

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
	moduleCalls map[string]*ModuleCall
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
	blockBuilder   BlockBuilder
	logger         zerolog.Logger
	isGraph        bool
	filteredBlocks []*Block
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
	logger zerolog.Logger,
	isGraph bool,
) *Evaluator {
	ctx := NewContext(&hcl.EvalContext{
		Functions: ExpFunctions(module.RootPath, logger),
	}, nil, logger)

	// Add any provider references from blocks in this module.
	// We do this here instead of loadModuleWithProviders to make sure
	// this also works with the root module
	if module.ProviderReferences == nil {
		module.ProviderReferences = make(map[string]*Block)
	}

	providerBlocks := module.Blocks.OfType("provider")
	for _, block := range providerBlocks {
		k := block.Label()

		alias := block.GetAttribute("alias")
		if alias != nil {
			k = k + "." + alias.AsString()
		}
		module.ProviderReferences[k] = block
	}

	for key, provider := range module.ProviderReferences {
		ctx.Set(provider.Values(), key)
	}

	if visitedModules == nil {
		visitedModules = make(map[string]map[string]cty.Value)
	}

	modulePath, err := filepath.Rel(module.RootPath, module.ModulePath)
	if err != nil {
		logging.Logger.Debug().Err(err).Str("rootPath", module.RootPath).
			Str("modulePath", module.ModulePath).Msg("Could not calculate relative path.module")

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

	l := logger.With().
		Str("module_name", moduleName).
		Str("module_source", module.Source).
		Str("module_path", module.ModulePath).
		Logger()

	return &Evaluator{
		module:         module,
		ctx:            ctx,
		inputVars:      inputVars,
		moduleMetadata: moduleMetadata,
		moduleCalls:    map[string]*ModuleCall{},
		visitedModules: visitedModules,
		workspace:      workspace,
		workingDir:     workingDir,
		blockBuilder:   blockBuilder,
		logger:         l,
		isGraph:        isGraph,
	}
}

func (e *Evaluator) AddFilteredBlocks(blocks ...*Block) {
	for _, block := range blocks {
		if block != nil {
			e.filteredBlocks = append(e.filteredBlocks, block)
		}
	}
}

// MissingVars returns a list of names of the variable blocks with missing input values.
func (e *Evaluator) MissingVars() []string {
	var missing []string

	blocks := e.module.Blocks.OfType("variable")
	for _, block := range blocks {
		name := block.Label()

		var sensitive bool
		value := block.GetAttribute("sensitive").Value()
		if !value.IsNull() {
			err := gocty.FromCtyValue(value, &sensitive)
			if err != nil {
				e.logger.Debug().Msgf("could not convert 'sensitive' attribute for variable.%s err: %s", name, err)
			}
		}

		if sensitive {
			continue
		}

		if sensitiveRegxp.MatchString(strings.ToLower(name)) {
			continue
		}

		_, v := e.evaluateVariable(block, e.inputVars)
		if v == errorNoVarValue {
			missing = append(missing, fmt.Sprintf("variable.%s", name))
		}
	}

	return missing
}

// Run builds the evaluation context using all provided blocks. It processes variables, locals,
// and early modules to populate the context with any values needed by other modules.
//
// After preparing the context, it expands all modules and resources, resolves references,
// and loads any additional modules that become available after expansion.
//
// The result is a fully evaluated root module that represents the complete Terraform configuration.
func (e *Evaluator) Run() (*Module, error) {
	var lastContext hcl.EvalContext

	// first we need to evaluate the top level Context - so this can be passed to any child modules that are found.
	e.logger.Debug().Msg("evaluating top level context")
	e.evaluate(lastContext)

	// let's load the modules now we have our top level context.
	e.loadModules(lastContext)
	e.logger.Debug().Msg("evaluating context after loading modules")
	e.evaluate(lastContext)

	// expand out resources and modules via count and evaluate again so that we can include
	// any module outputs and or count references.
	e.module.Blocks = e.expandBlocks(e.module.Blocks, lastContext)
	e.logger.Debug().Msg("evaluating context after expanding blocks")
	e.evaluate(lastContext)

	// returns all the evaluated Blocks under their given Module.
	return e.collectModules(), nil
}

func (e *Evaluator) collectModules() *Module {
	root := e.module
	for _, definition := range e.moduleCalls {
		root.Modules = append(root.Modules, definition.Module)
	}

	// Reload the provider references for this module instance
	// We need to do this so when we call ProviderConfigKey() we get the fully
	// resolved provider. We might be able to improve this by only evaluating the
	// provider block when we need it.
	for name, providerBlock := range root.ProviderReferences {
		e.ctx.Set(providerBlock.Values(), name)
	}

	if v := e.MissingVars(); len(v) > 0 {
		root.Warnings = append(root.Warnings, schema.NewDiagMissingVars(v...))
	}

	return &root
}

// evaluate runs a context evaluation loop until the context values are unchanged. We run this in a loop
// because variables can change because of outputs from other blocks in the context. Once all outputs have
// been evaluated and the context variables should remain unchanged. In reality 90% of cases will require
// 2 loops, however other complex modules will take > 2.
func (e *Evaluator) evaluate(lastContext hcl.EvalContext) {
	var i int

	for i = 0; i < maxContextIterations; i++ {
		e.evaluateStep(i)

		if reflect.DeepEqual(lastContext.Variables, e.ctx.Inner().Variables) {
			e.logger.Debug().Msg("evaluated outputs are the same a last context, exiting")
			break
		}

		if len(e.ctx.Inner().Variables) != len(lastContext.Variables) {
			lastContext.Variables = make(map[string]cty.Value, len(e.ctx.Inner().Variables))
		}

		for k, v := range e.ctx.Inner().Variables {
			lastContext.Variables[k] = v
		}
	}

	if i == maxContextIterations {
		e.logger.Debug().Msgf("hit max context iterations evaluating module %s", e.module.Name)
	}

}

// evaluateStep gets the values for all the Block types in the current Module that affect Context.
// It then sets these values on the Context so that they can be used in Block Attribute evaluation.
func (e *Evaluator) evaluateStep(i int) {
	e.logger.Debug().Msgf("starting context evaluation iteration %d", i+1)

	providers := e.getValuesByBlockType("provider")
	for key, provider := range providers.AsValueMap() {
		e.ctx.Set(provider, key)
	}

	e.ctx.Set(e.getValuesByBlockType("variable"), "var")
	e.ctx.Set(e.getValuesByBlockType("data"), "data")
	e.ctx.Set(e.getValuesByBlockType("locals"), "local")

	resources := e.getValuesByBlockType("resource")
	for key, resource := range resources.AsValueMap() {
		e.ctx.Set(resource, key)
	}

	e.ctx.Set(e.getValuesByBlockType("output"), "output")

	e.evaluateModules()
}

// evaluateModules loops over each of the moduleCalls in this Module and set a child Evaluator
// to run on the child Module Blocks. It passes the Evaluator the top level module Attributes as input variables.
func (e *Evaluator) evaluateModules() {
	for _, moduleCall := range e.moduleCalls {
		fullName := moduleCall.Definition.FullName()
		e.logger.Debug().Msgf("evaluating module call %s with source %s", fullName, moduleCall.Module.Source)
		vars := moduleCall.Definition.Values().AsValueMap()
		if oldVars, ok := e.visitedModules[fullName]; ok {
			if reflect.DeepEqual(vars, oldVars) {
				continue
			}

			e.logger.Debug().Msgf("module %s output vars have changed, evaluating again", fullName)
		}

		e.visitedModules[fullName] = vars

		moduleEvaluator := NewEvaluator(
			Module{
				Name:               fullName,
				Source:             moduleCall.Module.Source,
				Blocks:             moduleCall.Module.RawBlocks,
				RawBlocks:          moduleCall.Module.RawBlocks,
				RootPath:           e.module.RootPath,
				ModulePath:         moduleCall.Path,
				Modules:            nil,
				Parent:             &e.module,
				SourceURL:          moduleCall.Module.SourceURL,
				ProviderReferences: moduleCall.Module.ProviderReferences,
			},
			e.workingDir,
			vars,
			e.moduleMetadata,
			map[string]map[string]cty.Value{},
			e.workspace,
			e.blockBuilder,
			e.logger,
			e.isGraph,
		)

		moduleCall.Module, _ = moduleEvaluator.Run()
		outputs := moduleEvaluator.exportOutputs()
		if v := moduleCall.Module.Key(); v != nil {
			e.ctx.Set(outputs, "module", stripCount(moduleCall.Name), *v)
		} else if v := moduleCall.Module.Index(); v != nil {
			e.ctx.Set(outputs, "module", stripCount(moduleCall.Name), fmt.Sprintf("%d", *v))
		} else {
			e.ctx.Set(outputs, "module", moduleCall.Name)
		}
	}
}

// exportOutputs exports module outputs so that it can be used in Context evaluation.
func (e *Evaluator) exportOutputs() cty.Value {
	return e.module.Blocks.Outputs(false)
}

func (e *Evaluator) expandBlocks(blocks Blocks, lastContext hcl.EvalContext) Blocks {
	expanded := blocks

	var i int
	for i = 0; i < maxContextIterations; i++ {
		expanded = e.expandBlockForEaches(e.expandBlockCounts(expanded))

		if reflect.DeepEqual(lastContext.Variables, e.ctx.Inner().Variables) {
			e.logger.Debug().Msg("evaluated outputs are the same as prior evaluation, exiting and returning expanded block")
			break
		}

		if len(e.ctx.Inner().Variables) != len(lastContext.Variables) {
			lastContext.Variables = make(map[string]cty.Value, len(e.ctx.Inner().Variables))
		}

		for k, v := range e.ctx.Inner().Variables {
			lastContext.Variables[k] = v
		}
	}

	if i == maxContextIterations {
		e.logger.Debug().Msgf("hit max context iterations expanding blocks in module %s", e.module.Name)
	}

	return e.expandDynamicBlocks(expanded...)
}

func (e *Evaluator) expandDynamicBlocks(blocks ...*Block) Blocks {
	var newBlocks Blocks

	for _, b := range blocks {
		if b.HasDynamicBlock() {
			newBlocks = append(newBlocks, e.expandDynamicBlock(b))
			continue
		}

		newBlocks = append(newBlocks, b)
	}

	return newBlocks
}

func (e *Evaluator) expandDynamicBlock(b *Block) *Block {
	var newChildBlocks Blocks
	for _, child := range b.Children() {
		// if the child block is not a dynamic there is nothing to do
		// so add it to the newBlocks straight away and continue.
		if child.Type() != "dynamic" {
			newChildBlocks = append(newChildBlocks, child)
			continue
		}

		e.logger.Debug().Msgf("expanding block %s because a dynamic block was found %s", b.LocalName(), child.LocalName())
		blockName := child.TypeLabel()

		// for each expanded dynamic block add the generated "content" as a
		// new child block into the parent block.
		expanded := e.expandBlockForEaches([]*Block{child})
		for i, ex := range expanded {
			content := ex.GetChildBlock("content")
			if content == nil {
				continue
			}

			_ = e.expandDynamicBlocks(content)

			content.SetLabels([]string{})
			content.SetType(blockName)

			for attrName, attr := range content.AttributesAsMap() {
				b.context.Root().SetByDot(attr.Value(), fmt.Sprintf("%s.%s.%d.%s", b.Reference().String(), blockName, i, attrName))
			}

			newChildBlocks = append(newChildBlocks, content)
		}
	}

	for i, block := range newChildBlocks {
		newChildBlocks[i] = e.expandDynamicBlock(block)
	}

	return &Block{
		HCLBlock:    b.HCLBlock,
		UniqueAttrs: b.UniqueAttrs,
		context:     b.context,
		moduleBlock: b.moduleBlock,
		rootPath:    b.rootPath,
		expanded:    b.expanded,
		cloneIndex:  b.cloneIndex,
		original:    b.original,
		childBlocks: newChildBlocks,
		parent:      b.parent,
		verbose:     b.verbose,
		logger:      b.logger,
		isGraph:     b.isGraph,
		newMock:     b.newMock,
		attributes:  b.attributes,
		reference:   b.reference,
		Filename:    b.Filename,
		StartLine:   b.StartLine,
		EndLine:     b.EndLine,
	}
}

// expandBlockForEaches expands the block for_each attributes. Every block that is expanded has it's value
// reset in the context map. The original value of the block is replaced with the expanded values.
// This is required as otherwise we'll have the original attributes polluting the expanded values.
// Leaving the original attributes is problematic as any resources referencing the expanded block will
// iterate over those as well. So the context map will get overwritten as follows:
//
// value of:
//
//	test.test
//		id: test
//		arn: test-arn
//
// that gets expanded should be:
//
//	test.test
//		a:
//		  id: test
//		  arn: test-arn
//		b:
//		  id: test
//		  arn: test-arn
//
// and not:
//
//	   test.test
//			id: test
//			arn: test-arn
//			a:
//			  id: test
//			...
//
// which means that blocks written like:
//
//	resource "aws_eip" "nat_gateway" {
//		for_each   = test.test
//			...
//	}
//
// will expand correctly to: aws_eip.nat_gateway[a], aws_eip.nat_gateway[b]. Rather than aws_eip.nat_gateway[id], aws_eip.nat_gateway[a] ...
func (e *Evaluator) expandBlockForEaches(blocks Blocks) Blocks {
	var expanded Blocks
	var haveChanged = make(map[string]*Block)

	for _, block := range blocks {
		forEachAttr := block.GetAttribute("for_each")
		if forEachAttr == nil {
			expanded = append(expanded, block)
			continue
		}

		if !e.isGraph && (block.IsCountExpanded() || !block.IsForEachReferencedExpanded(blocks) || !shouldExpandBlock(block)) {
			original := block.original.GetAttribute("for_each")
			if !original.HasChanged() {
				expanded = append(expanded, block)
			} else {
				haveChanged[block.original.FullName()] = block.original
			}

			continue
		}

		var iterName string
		if block.Type() == "dynamic" {
			iterName = e.getIteratorAttrName(block)
		}

		e.logger.Debug().Msgf("expanding block %s because a for_each attribute was found", block.LocalName())

		value := forEachAttr.Value()
		if !value.IsNull() && value.IsKnown() && forEachAttr.IsIterable() {
			labels := block.Labels()
			if block.Type() != "resource" {
				labels = append([]string{block.Type()}, labels...)
			}

			// set the context to an empty object so that we don't have any of the prior unexpanded attributes
			// like id/arn e.t.c polluting the expansion. Otherwise we can end up with lots of expanded resources
			// with attribute names e.g. aws_instance.test[id], aws_instance.test[arn].
			e.ctx.Set(cty.ObjectVal(make(map[string]cty.Value)), labels...)

			value.ForEachElement(func(key cty.Value, val cty.Value) bool {
				clone := e.blockBuilder.CloneBlock(block, key)

				ctx := clone.Context()

				var keyStr string
				err := gocty.FromCtyValue(key, &keyStr)
				if err != nil {
					e.logger.Debug().Err(err).Msgf("could not marshal gocty key %s to string", key)
				}

				ctx.SetByDot(key, "each.key")
				ctx.SetByDot(val, "each.value")
				// If we have an iterName let's also set a child context with this as a prefix.
				// This enables proper support for dynamic block content generation. Dynamic
				// blocks can use different names outside the regular "each." to refer to
				// attributes in the loop expansion see [Evaluator.getIteratorAttrName] for more
				// info.
				if iterName != "" {
					ctx.SetByDot(key, iterName+".key")
					ctx.SetByDot(val, iterName+".value")
				}

				ctx.Set(key, block.TypeLabel(), "key")
				ctx.Set(val, block.TypeLabel(), "value")

				cloneValues := clone.Values()
				e.ctx.Set(cloneValues, append(labels, keyStr)...)

				if clone.Type() == "module" {
					if v, ok := e.moduleCalls[clone.FullName()]; ok {
						v.Definition = clone
					} else {
						modCall, err := e.loadModuleWithProviders(clone)
						if err != nil {
							e.logger.Debug().Err(err).Msgf("failed to create expanded module call, could not load module %s", clone.FullName())
							return false
						}
						e.moduleCalls[clone.FullName()] = modCall
					}
				}

				expanded = append(expanded, clone)

				return false
			})

			if block.Type() == "module" {
				e.logger.Debug().Msgf("deleting module from moduleCalls since it has been expanded %s", block.FullName())
				delete(e.moduleCalls, block.FullName())
			}
		} else {
			expanded = append(expanded, block)
		}
	}

	if len(haveChanged) > 0 {
		var changes = make(Blocks, 0, len(haveChanged))
		for _, block := range haveChanged {
			changes = append(changes, block)
		}

		eaches := e.expandBlockForEaches(changes)
		return append(expanded, eaches...)
	}

	return expanded
}

// getIteratorAttrName returns the iterator key which is used to set the child
// context of a dynamic block. Normally, in the context of a for_each expansion
// we set the child context to have a prefix of "each". This allows correct
// evaluation of attributes referencing "each.value." or "each.key.". However,
// dynamic blocks are special cases. Not only can they use "each" in the content block,
// but they also are able to refer to the dynamic block label, e.g:
//
//	 dynamic "setting" {
//	   for_each = var.settings
//	   content {
//	     namespace = setting.value.namespace.
//			...
//	   }
//	 }
//
// additionally they have a "iterator" block:
// https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks
// which allow the user to specify a different prefix for a for_each expansion.
// For example:
//
//	  dynamic "foo" {
//	   for_each = {
//	      ...
//	   }
//	   iterator = device
//	   content {
//	     name = device.value.name
//			...
//	   }
//	 }
//
// In this case we need to return the iterator name ("device" in the above
// example) so that we properly set the child context with the correct prefix.
// Terraform expects the iterator key to be specified by a root name traversal
// here (note the lack of quotes around the device value for iterator in the
// above example). This means we can't just do `attr.Value().AsString()` as
// normal because the actual value returns `UnknownType`. Instead, we need to
// "pop" the RootName off the expression.
func (e *Evaluator) getIteratorAttrName(block *Block) string {
	iterator := block.GetAttribute("iterator")
	if iterator != nil {
		travers, diags := hcl.AbsTraversalForExpr(iterator.HCLAttr.Expr)
		if diags.HasErrors() {
			e.logger.Debug().Err(diags).Msg("failed to get abs traversal for dynamic block iterator attr")
			return ""
		}

		if len(travers) != 1 {
			e.logger.Debug().Msg("dynamic block iterator had incorrect expression length, unable to retrieve iterator name")
			return ""
		}

		return travers.RootName()
	}

	return block.TypeLabel()
}

func shouldExpandBlock(block *Block) bool {
	_, isValidType := validBlocksToExpand[block.Type()]
	return isValidType
}

func (e *Evaluator) expandBlockCounts(blocks Blocks) Blocks {
	var expanded Blocks
	var haveChanged = make(map[string]*Block)
	for _, block := range blocks {
		countAttr := block.GetAttribute("count")
		if countAttr == nil {
			expanded = append(expanded, block)
			continue
		}

		if !e.isGraph && (block.IsCountExpanded() || !shouldExpandBlock(block)) {
			original := block.original.GetAttribute("count")
			if !original.HasChanged() {
				expanded = append(expanded, block)
			} else {
				haveChanged[block.original.FullName()] = block.original
			}

			continue
		}

		count := 1
		value := countAttr.Value()
		if !value.IsNull() && value.IsKnown() {
			v := countAttr.AsInt()
			if v >= 0 && v <= math.MaxInt32 {
				count = int(v)
			}
		}

		e.logger.Debug().Msgf("expanding block %s because a count attribute of value %d was found", block.LocalName(), count)

		vals := make([]cty.Value, count)
		for i := 0; i < count; i++ {
			c, _ := gocty.ToCtyValue(i, cty.Number)
			clone := e.blockBuilder.CloneBlock(block, c)

			expanded = append(expanded, clone)
			vals[i] = clone.Values()
		}

		e.ctx.SetByDot(cty.TupleVal(vals), block.Reference().String())
	}

	if len(haveChanged) > 0 {
		var changes = make(Blocks, 0, len(haveChanged))
		for _, block := range haveChanged {
			changes = append(changes, block)
		}

		counts := e.expandBlockCounts(changes)
		return append(expanded, counts...)
	}

	return expanded
}

func (e *Evaluator) evaluateVariable(b *Block, inputVars map[string]cty.Value) (cty.Value, error) {
	if b.Label() == "" {
		return cty.DynamicVal, fmt.Errorf("empty label - cannot resolve")
	}

	attributes := b.AttributesAsMap()
	if attributes == nil {
		return cty.DynamicVal, fmt.Errorf("cannot resolve variable with no attributes")
	}

	attrType := attributes["type"]
	if override, exists := inputVars[b.Label()]; exists {
		val, err := e.convertType(b, override, attrType)
		if err == nil {
			return val, nil
		}

		return override, nil
	}

	if def, exists := attributes["default"]; exists {
		val, err := e.convertType(b, def.Value(), attrType)
		if err == nil {
			return val, nil
		}
	}

	c, err := e.convertType(b, cty.DynamicVal, attrType)
	if err != nil {
		return c, err
	}

	return c, errorNoVarValue
}

func (e *Evaluator) convertType(b *Block, val cty.Value, attrType *Attribute) (cty.Value, error) {
	if attrType == nil || val.IsNull() || !val.IsKnown() {
		return val, nil
	}

	ty, def, diag := typeexpr.TypeConstraintWithDefaults(attrType.HCLAttr.Expr)
	if diag.HasErrors() {
		e.logger.Debug().Err(diag).Msgf("error trying to convert variable %s to type %s", b.Label(), attrType.AsString())
		return val, nil
	}

	// Check if default values exist for the variable type definition. If they do, we
	// merge these defaults with any existing values. This ensures that variables
	// with optional types that have default values e.g., optional(string, "foo")
	// are fully resolved.
	if def != nil {
		val = def.Apply(val)
	}
	return convert.Convert(val, ty)
}

func (e *Evaluator) evaluateOutput(b *Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.DynamicVal, fmt.Errorf("empty label - cannot resolve")
	}

	attribute := b.GetAttribute("value")
	if attribute == nil {
		return cty.DynamicVal, fmt.Errorf("cannot resolve variable with no attributes")
	}
	return attribute.Value(), nil
}

func (e *Evaluator) getValuesByBlockType(blockType string) cty.Value {
	blocksOfType := e.module.Blocks.OfType(blockType)
	values := make(map[string]cty.Value)

	for _, b := range blocksOfType {
		switch b.Type() {
		case "variable": // variables are special in that their value comes from the "default" attribute
			val, err := e.evaluateVariable(b, e.inputVars)
			if err != nil {
				e.logger.Debug().Err(err).Msgf("could not evaluate variable %s ignoring", b.FullName())
				continue
			}

			e.logger.Debug().Msgf("adding variable %s to the evaluation context", b.Label())
			values[b.Label()] = val
		case "output":
			val, err := e.evaluateOutput(b)
			if err != nil {
				e.logger.Debug().Err(err).Msgf("could not evaluate output %s ignoring", b.FullName())
				continue
			}

			e.logger.Debug().Msgf("adding output %s to the evaluation context", b.Label())
			values[b.Label()] = val
		case "locals":
			for key, val := range b.Values().AsValueMap() {
				e.logger.Debug().Msgf("adding local %s to the evaluation context", key)

				values[key] = val
			}
		case "provider":
			provider := b.Label()
			if provider == "" {
				continue
			}

			values[provider] = e.evaluateProvider(b, values)
		case "module":
			if b.Label() == "" {
				continue
			}

			e.logger.Debug().Msgf("adding %s %s to the evaluation context", b.Type(), b.Label())
			values[b.Label()] = b.Values()
		case "resource", "data":
			if len(b.Labels()) < 2 {
				continue
			}

			e.logger.Debug().Msgf("adding %s %s to the evaluation context", b.Type(), b.Label())
			values[b.Labels()[0]] = e.evaluateResourceOrData(b, values)
		}

	}

	return cty.ObjectVal(values)
}

// evaluateProvider evaluates a provider block.
// The values map is used to pass in the current context values. This is only needed
// for the legacy evaluator and is not used for the graph evaluator.
func (e *Evaluator) evaluateProvider(b *Block, values map[string]cty.Value) cty.Value {
	provider := b.Label()
	v, exists := values[provider]

	alias := b.GetAttribute("alias")
	if alias == nil && exists {
		return mergeObjects(v, b.Values())
	}

	if alias == nil {
		return b.Values()
	}

	var str string
	err := gocty.FromCtyValue(alias.Value(), &str)
	if err != nil {
		return cty.ObjectVal(values)
	}

	if !exists {
		return cty.ObjectVal(map[string]cty.Value{
			str: b.Values(),
		})
	}

	ob := v.AsValueMap()
	if ob == nil {
		ob = make(map[string]cty.Value)
	}
	ob[str] = b.Values()
	return cty.ObjectVal(ob)
}

// evaluateResourceOrData evaluates a resource or data block.
// The values map is used to pass in the current context values. This is only needed
// for the legacy evaluator and is not used for the graph evaluator.
func (e *Evaluator) evaluateResourceOrData(b *Block, values map[string]cty.Value) cty.Value {
	labels := b.Labels()

	blockMap, ok := values[labels[0]]
	if !ok {
		if values == nil {
			values = make(map[string]cty.Value)
		}

		values[labels[0]] = cty.ObjectVal(make(map[string]cty.Value))
		blockMap = values[labels[0]]
	}

	valueMap := blockMap.AsValueMap()
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}

	if k := b.Key(); k != nil {
		e.logger.Debug().Msgf("expanding block %s to be available for for_each key %s", b.FullName(), *k)
		valueMap[stripCount(labels[1])] = e.expandedEachBlockToValue(b, valueMap)
		return cty.ObjectVal(valueMap)
	}

	if k := b.Index(); k != nil {
		e.logger.Debug().Msgf("expanding block %s to be available for index key %d", b.FullName(), *k)
		valueMap[stripCount(labels[1])] = expandCountBlockToValue(b, valueMap)
		return cty.ObjectVal(valueMap)
	}

	valueMap[b.Labels()[1]] = b.Values()
	return cty.ObjectVal(valueMap)
}

func expandCountBlockToValue(b *Block, existingValues map[string]cty.Value) cty.Value {
	k := b.Index()
	if k == nil {
		return cty.DynamicVal
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

func (e *Evaluator) expandedEachBlockToValue(b *Block, existingValues map[string]cty.Value) cty.Value {
	k := b.Key()
	if k == nil {
		return cty.DynamicVal
	}

	ob := make(map[string]cty.Value)

	name := b.Labels()[1]
	eachMap := existingValues[stripCount(name)]
	if !eachMap.IsNull() && eachMap.IsKnown() {
		if !eachMap.Type().IsObjectType() && !eachMap.Type().IsMapType() {
			e.logger.Debug().Str(
				"block", b.Label(),
			).Msgf("skipping unexpected cty value type '%s' for existing for_each context value", eachMap.GoString())

			ob[*k] = b.Values()
			return cty.ObjectVal(ob)
		}

		for ek, v := range eachMap.AsValueMap() {
			ob[ek] = v
		}
	}

	ob[*k] = b.Values()
	return cty.ObjectVal(ob)
}

// loadModuleWithProviders takes in a module "x" {} block and loads resources etc. into e.moduleBlocks.
// Additionally, it returns variables to add to ["module.x.*"] variables
func (e *Evaluator) loadModuleWithProviders(b *Block) (*ModuleCall, error) {
	modCall, err := e.loadModule(b)
	if err != nil {
		return modCall, err
	}

	// Pass any provider references that should be inherited by the module.
	// This includes any implicit providers that are inherited from the parent
	// module, as well as any explicit provider references that are passed in
	// via the "providers" attribute.
	providerRefs := map[string]*Block{}

	for key, block := range e.module.ProviderReferences {
		providerRefs[key] = block
	}

	providerAttr := b.GetAttribute("providers")
	if providerAttr != nil {
		decodedProviders := providerAttr.DecodeProviders()
		for key, val := range decodedProviders {
			providerRefs[key] = providerRefs[val]
		}
	}

	modCall.Module.ProviderReferences = providerRefs

	return modCall, nil
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
		}
	}

	if source == "" {
		return nil, fmt.Errorf("could not read module source attribute at %s", b.FullName())
	}

	var modulePath string
	var moduleURL string

	if e.moduleMetadata != nil {
		// if we have module metadata we can parse all the modules as they'll be cached locally!

		// Strip any "module." and "[*]" parts from the module name so it matches the manifest key format
		key := modReplace.ReplaceAllString(b.FullName(), "")
		key = nestedModReplace.ReplaceAllString(key, ".")
		key = modArrayPartReplace.ReplaceAllString(key, "")

		module := e.moduleMetadata.Get(key)
		modulePath = module.Dir
		moduleURL = module.URL()
		e.logger.Debug().Msgf("using path '%s' for module '%s' based on key '%s'", modulePath, b.FullName(), key)
	}

	if modulePath == "" {
		if !strings.HasPrefix(source, fmt.Sprintf(".%c", os.PathSeparator)) && !strings.HasPrefix(source, fmt.Sprintf("..%c", os.PathSeparator)) {
			reg := "registry.terraform.io/" + source
			return nil, fmt.Errorf("missing module with source '%s %s' -  try to 'terraform init' first", reg, source)
		}

		// combine the current calling module with relative source of the module
		modulePath = filepath.Join(e.module.ModulePath, source)
	}

	blocks, err := e.blockBuilder.BuildModuleBlocks(b, modulePath, e.module.RootPath)
	if err != nil {
		return nil, err
	}
	e.logger.Debug().Msgf("loaded module '%s' (requested at %s)", modulePath, b.FullName())

	return &ModuleCall{
		Name:       b.Label(),
		Path:       modulePath,
		Definition: b,
		Module: &Module{
			Name:       b.TypeLabel(),
			Source:     source,
			SourceURL:  moduleURL,
			Blocks:     blocks,
			RawBlocks:  blocks,
			RootPath:   e.module.RootPath,
			ModulePath: modulePath,
			Parent:     &e.module,
		},
	}, nil
}

// loadAndEvaluateEarlyModules loads early module calls and evaluates them.
func (e *Evaluator) loadAndEvaluateEarlyModules(lastContext hcl.EvalContext) {
	e.logger.Debug().Msg("loading early module calls")

	var earlyModuleBlocks Blocks
	var filtered Blocks

	for _, block := range e.module.Blocks {
		if block.Type() == "module" && block.Label() == "with_output" {
			earlyModuleBlocks = append(earlyModuleBlocks, block)
		} else {
			filtered = append(filtered, block)
		}
	}

	// Load and evaluate early modules without expansion
	for _, moduleBlock := range earlyModuleBlocks {
		moduleCall, err := e.loadModuleWithProviders(moduleBlock)
		if err != nil {
			e.logger.Debug().Err(err).Msgf("failed to load early module %s", moduleBlock.LocalName())
			continue
		}
		e.moduleCalls[moduleBlock.FullName()] = moduleCall

		// Evaluate it
		moduleEvaluator := NewEvaluator(
			Module{
				Name:               moduleCall.Definition.FullName(),
				Source:             moduleCall.Module.Source,
				Blocks:             moduleCall.Module.RawBlocks,
				RawBlocks:          moduleCall.Module.RawBlocks,
				RootPath:           e.module.RootPath,
				ModulePath:         moduleCall.Path,
				Modules:            nil,
				Parent:             &e.module,
				SourceURL:          moduleCall.Module.SourceURL,
				ProviderReferences: moduleCall.Module.ProviderReferences,
			},
			e.workingDir,
			moduleCall.Definition.Values().AsValueMap(),
			e.moduleMetadata,
			map[string]map[string]cty.Value{},
			e.workspace,
			e.blockBuilder,
			e.logger,
			e.isGraph,
		)

		moduleCall.Module, _ = moduleEvaluator.Run()
		outputs := moduleEvaluator.exportOutputs()

		// Inject outputs into context
		e.ctx.Set(outputs, "module", moduleCall.Name)
	}

	// Update the top-level blocks (early modules + rest)
	e.module.Blocks = append(filtered, earlyModuleBlocks...)
}

// loadModules reads all module blocks and loads the underlying modules, adding blocks to moduleCalls.
func (e *Evaluator) loadModules(lastContext hcl.EvalContext) {
	e.logger.Debug().Msg("loading module calls")

	var moduleBlocks Blocks
	var filtered Blocks

	for _, block := range e.module.Blocks {
		if block.Type() == "module" {
			sourceAttr := block.GetAttribute("source")
			evaluated := false

			if sourceAttr != nil && sourceAttr.Value().IsKnown() && sourceAttr.Value().Type() == cty.String {
				// Only evaluate early if:
				// - all attributes are known
				// - block does NOT use count or for_each
				hasCount := block.GetAttribute("count") != nil
				hasForEach := block.GetAttribute("for_each") != nil

				if !e.blockHasUnknowns(block) && !hasCount && !hasForEach {
					moduleCall, err := e.loadModuleWithProviders(block)
					if err != nil {
						e.logger.Debug().Err(err).Msgf("failed to evaluate module %s", block.FullName())
					} else {
						e.moduleCalls[block.FullName()] = moduleCall

						moduleEvaluator := NewEvaluator(
							Module{
								Name:               moduleCall.Definition.FullName(),
								Source:             moduleCall.Module.Source,
								Blocks:             moduleCall.Module.RawBlocks,
								RawBlocks:          moduleCall.Module.RawBlocks,
								RootPath:           e.module.RootPath,
								ModulePath:         moduleCall.Path,
								Modules:            nil,
								Parent:             &e.module,
								SourceURL:          moduleCall.Module.SourceURL,
								ProviderReferences: moduleCall.Module.ProviderReferences,
							},
							e.workingDir,
							moduleCall.Definition.Values().AsValueMap(),
							e.moduleMetadata,
							map[string]map[string]cty.Value{},
							e.workspace,
							e.blockBuilder,
							e.logger,
							e.isGraph,
						)

						moduleCall.Module, _ = moduleEvaluator.Run()
						outputs := moduleEvaluator.exportOutputs()
						e.ctx.Set(outputs, "module", moduleCall.Name)

						evaluated = true
					}
				}
			}

			// If not evaluated, queue for later expansion
			if !evaluated {
				moduleBlocks = append(moduleBlocks, block)
			}
		} else {
			// Keep all non-module blocks
			filtered = append(filtered, block)
		}
	}

	// Expand module blocks with known count/for_each
	expanded := e.expandBlocks(moduleBlocks.SortedByCaller(), lastContext)
	filtered = append(filtered, expanded...)
	e.module.Blocks = filtered

	// Load the newly expanded module instances
	for _, moduleBlock := range expanded {
		if moduleBlock.Label() == "" {
			continue
		}

		moduleCall, err := e.loadModuleWithProviders(moduleBlock)
		if err != nil {
			e.logger.Debug().Err(err).Msgf("failed to load module %s ignoring", moduleBlock.FullName())
			continue
		}

		e.moduleCalls[moduleBlock.FullName()] = moduleCall
	}
}

func (e *Evaluator) blockHasUnknowns(block *Block) bool {
	for _, attr := range block.GetAttributes() {
		val := attr.Value()
		if !val.IsKnown() {
			e.logger.Debug().Msgf("attribute %s in block %s is not known", attr.Name(), block.FullName())
			return true
		}
	}
	return false
}

// ExpFunctions returns the set of functions that should be used to when evaluating
// expressions in the receiving scope.
func ExpFunctions(baseDir string, logger zerolog.Logger) map[string]function.Function {
	fns := map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          filesystem.AbsPathFunc,
		"basename":         filesystem.BasenameFunc,
		"base64decode":     encoding.Base64DecodeFunc,
		"base64encode":     encoding.Base64EncodeFunc,
		"base64gzip":       componentsFuncs.Base64GzipFunc,
		"base64sha256":     componentsFuncs.Base64Sha256Func,
		"base64sha512":     componentsFuncs.Base64Sha512Func,
		"bcrypt":           crypto.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             stdlib.CeilFunc,
		"chomp":            stdlib.ChompFunc,
		"cidrhost":         componentsFuncs.CidrHostFunc,
		"cidrnetmask":      cidr.NetmaskFunc,
		"cidrsubnet":       componentsFuncs.CidrSubnetFunc,
		"cidrsubnets":      cidr.SubnetsFunc,
		"coalesce":         funcs.CoalesceFunc, // customized from stdlib
		"coalescelist":     stdlib.CoalesceListFunc,
		"compact":          stdlib.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         stdlib.ContainsFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          filesystem.DirnameFunc,
		"distinct":         stdlib.DistinctFunc,
		"element":          stdlib.ElementFunc,
		"endswith":         componentsFuncs.EndsWithFunc,
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
		"formatdate":       funcs.FormatDateFunc, // wraps stdlib.FormatDateFunc
		"formatlist":       stdlib.FormatListFunc,
		"indent":           stdlib.IndentFunc,
		"index":            componentsFuncs.IndexFunc,
		"join":             stdlib.JoinFunc,
		"jsondecode":       funcs.JSONDecodeFunc, // customized from stdlib to handle mocks
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             stdlib.KeysFunc,
		"length":           componentsFuncs.LengthFunc,
		"list":             componentsFuncs.ListFunc,
		"log":              stdlib.LogFunc,
		"lookup":           componentsFuncs.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"map":              componentsFuncs.MapFunc,
		"matchkeys":        componentsFuncs.MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              crypto.Md5Func,
		"merge":            funcs.MergeFunc, // customized from stdlib
		"min":              stdlib.MinFunc,
		"parseint":         stdlib.ParseIntFunc,
		"pathexpand":       filesystem.PathExpandFunc,
		"infracostlog":     funcs.LogArgs(logger), // custom
		"infracostprint":   funcs.PrintArgs,       // custom
		"pow":              stdlib.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          componentsFuncs.ReplaceFunc,
		"reverse":          stdlib.ReverseListFunc,
		"rsadecrypt":       componentsFuncs.RsaDecryptFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       stdlib.SetProductFunc,
		"setsubtract":      stdlib.SetSubtractFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             crypto.Sha1Func,
		"sha256":           crypto.Sha256Func,
		"sha512":           crypto.Sha512Func,
		"signum":           stdlib.SignumFunc,
		"slice":            stdlib.SliceFunc,
		"sort":             stdlib.SortFunc,
		"split":            stdlib.SplitFunc,
		"startswith":       componentsFuncs.StartsWithFunc,
		"strcontains":      componentsFuncs.StrContainsFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"timestamp":        funcs.MockTimestampFunc, // custom. We want to return a deterministic value each time
		"timeadd":          stdlib.TimeAddFunc,
		"title":            stdlib.TitleFunc,
		"tostring":         funcs.MakeToFunc(cty.String),
		"tonumber":         funcs.MakeToFunc(cty.Number),
		"tobool":           funcs.MakeToFunc(cty.Bool),
		"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"transpose":        componentsFuncs.TransposeFunc,
		"trim":             stdlib.TrimFunc,
		"trimprefix":       stdlib.TrimPrefixFunc,
		"trimspace":        stdlib.TrimSpaceFunc,
		"trimsuffix":       stdlib.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        encoding.URLEncodeFunc,
		"uuid":             componentsFuncs.UUIDFunc,
		"uuidv5":           componentsFuncs.UUIDV5Func,
		"values":           stdlib.ValuesFunc,
		"yamldecode":       funcs.YAMLDecodeFunc, // customized yaml.YAMLDecodeFunc
		"yamlencode":       yaml.YAMLEncodeFunc,
		"zipmap":           stdlib.ZipmapFunc,
	}

	fns["templatefile"] = funcs.MakeTemplateFileFunc(baseDir, func() map[string]function.Function {
		return fns
	})

	return fns
}
