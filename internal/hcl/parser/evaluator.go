package parser

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/block"
)

const maxContextIterations = 32

type visitedModule struct {
	name                string
	path                string
	definitionReference string
}

type Evaluator struct {
	ctx               *block.Context
	blocks            block.Blocks
	moduleDefinitions []*ModuleDefinition
	visitedModules    []*visitedModule
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
	blocks block.Blocks,
	inputVars map[string]cty.Value,
	moduleMetadata *ModulesMetadata,
	visitedModules []*visitedModule,
	stopOnHCLError bool,
	workspace string,
) *Evaluator {
	ctx := block.NewContext(&hcl.EvalContext{
		Functions: Functions(modulePath),
	}, nil)

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
		if visited := func(module *ModuleDefinition) bool {
			for _, v := range e.visitedModules {
				if v.name == module.Name && v.path == module.Path && module.Definition.Reference().String() == v.definitionReference {
					log.Debugf("Module [%s:%s:%s] has already been seen", v.name, v.path, v.definitionReference)
					return true
				}
			}
			return false
		}(module); visited {
			continue
		}

		e.visitedModules = append(e.visitedModules, &visitedModule{module.Name, module.Path, module.Definition.Reference().String()})

		vars := module.Definition.Values().AsValueMap()
		moduleEvaluator := NewEvaluator(e.projectRootPath, module.Path, e.workingDir, module.Modules[0].GetBlocks(), vars, e.moduleMetadata, e.visitedModules, e.stopOnHCLError, e.workspace)
		module.Modules, _ = moduleEvaluator.EvaluateAll()
		// export module outputs
		e.ctx.Set(moduleEvaluator.ExportOutputs(), "module", module.Name)
	}
}

// export module outputs to a parent
func (e *Evaluator) ExportOutputs() cty.Value {
	data := make(map[string]cty.Value)
	for _, block := range e.blocks.OfType("output") {
		attr := block.GetAttribute("value")
		if attr.IsNil() {
			continue
		}
		data[block.Label()] = attr.Value()
	}
	return cty.ObjectVal(data)
}

func (e *Evaluator) EvaluateAll() ([]block.Module, error) {
	var lastContext hcl.EvalContext
	for i := 0; i < maxContextIterations; i++ {
		e.evaluateStep(i)

		// if ctx matches the last evaluation, we can bail, nothing left to resolve
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

	log.Debug("Loading modules...")
	e.moduleDefinitions = e.loadModules(true)

	// expand out resources and modules via count
	e.blocks = e.expandBlocks(e.blocks)

	for i := 0; i < maxContextIterations; i++ {
		e.evaluateStep(i)

		// if ctx matches the last evaluation, we can bail, nothing left to resolve
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

	var modules []block.Module
	modules = append(modules, block.NewHCLModule(e.projectRootPath, e.modulePath, e.blocks))
	for _, definition := range e.moduleDefinitions {
		modules = append(modules, definition.Modules...)
	}

	return modules, nil
}

func (e *Evaluator) expandBlocks(blocks block.Blocks) block.Blocks {
	return e.expandDynamicBlocks(e.expandBlockForEaches(e.expandBlockCounts(blocks))...)
}

func (e *Evaluator) expandDynamicBlocks(blocks ...block.Block) block.Blocks {
	for _, b := range blocks {
		e.expandDynamicBlock(b)
	}
	return blocks
}

func (e *Evaluator) expandDynamicBlock(b block.Block) {
	for _, sub := range b.AllBlocks() {
		e.expandDynamicBlock(sub)
	}
	for _, sub := range b.AllBlocks().OfType("dynamic") {
		blockName := sub.TypeLabel()
		expanded := e.expandBlockForEaches([]block.Block{sub})
		for _, ex := range expanded {
			if content := ex.GetBlock("content"); content.IsNotNil() {
				_ = e.expandDynamicBlocks(content)
				b.InjectBlock(content, blockName)
			}
		}
	}
}

func (e *Evaluator) expandBlockForEaches(blocks block.Blocks) block.Blocks {
	var forEachFiltered block.Blocks
	for _, block := range blocks {
		forEachAttr := block.GetAttribute("for_each")
		if forEachAttr.IsNil() || block.IsCountExpanded() || (block.Type() != "resource" && block.Type() != "module" && block.Type() != "dynamic") {
			forEachFiltered = append(forEachFiltered, block)
			continue
		}
		if !forEachAttr.Value().IsNull() && forEachAttr.Value().IsKnown() && forEachAttr.IsIterable() {
			forEachAttr.Each(func(key cty.Value, val cty.Value) {
				clone := block.Clone(key)

				ctx := clone.Context()

				e.copyVariables(block, clone)

				ctx.SetByDot(key, "each.key")
				ctx.SetByDot(val, "each.value")

				ctx.Set(key, block.TypeLabel(), "key")
				ctx.Set(val, block.TypeLabel(), "value")

				log.Debugf("Added %s from for_each", clone.Reference())
				forEachFiltered = append(forEachFiltered, clone)
			})
		}
	}

	return forEachFiltered
}

func (e *Evaluator) expandBlockCounts(blocks block.Blocks) block.Blocks {
	var countFiltered block.Blocks
	for _, block := range blocks {
		countAttr := block.GetAttribute("count")
		if countAttr.IsNil() || block.IsCountExpanded() || (block.Type() != "resource" && block.Type() != "module") {
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

func (e *Evaluator) copyVariables(from, to block.Block) {

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

func (e *Evaluator) evaluateVariable(b block.Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.NilVal, fmt.Errorf("empty label - cannot resolve")
	}

	attributes := b.Attributes()
	if attributes == nil {
		return cty.NilVal, fmt.Errorf("cannot resolve variable with no attributes")
	}

	if override, exists := e.inputVars[b.Label()]; exists {
		return override, nil
	} else if def, exists := attributes["default"]; exists {
		return def.Value(), nil
	}

	return cty.NilVal, fmt.Errorf("no value found")
}

func (e *Evaluator) evaluateOutput(b block.Block) (cty.Value, error) {
	if b.Label() == "" {
		return cty.NilVal, fmt.Errorf("empty label - cannot resolve")
	}

	attribute := b.GetAttribute("value")
	if attribute.IsNil() {
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
