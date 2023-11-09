package hcl

import (
	"fmt"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

func moduleCallID(moduleAddress string) string {
	return fmt.Sprintf("call:%s", moduleAddress)
}

type VertexModuleCall struct {
	logger        *logrus.Entry
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexModuleCall) ID() string {
	return moduleCallID(v.block.FullName())
}

func (v *VertexModuleCall) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexModuleCall) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexModuleCall) Visit(mutex *sync.Mutex) error {
	moduleInstances := v.moduleConfigs.Get(v.block.ModuleAddress())
	if len(moduleInstances) == 0 {
		return fmt.Errorf("no module instances found for module address %q", v.block.ModuleAddress())
	}

	for _, moduleInstance := range moduleInstances {
		e := moduleInstance.evaluator
		blockInstance := e.module.Blocks.FindLocalName(v.block.LocalName())

		if blockInstance == nil {
			return fmt.Errorf("could not find block %q in module %q", v.ID(), moduleInstance.name)
		}

		err := v.evaluate(e, blockInstance, mutex)
		if err != nil {
			return fmt.Errorf("could not evaluate module %q", v.block.FullName())
		}

		expanded, err := v.expand(e, blockInstance, mutex)
		if err != nil {
			return fmt.Errorf("could not expand module %q", v.block.FullName())
		}

		e.AddFilteredBlocks(expanded...)
	}

	return nil
}

func (v *VertexModuleCall) evaluate(e *Evaluator, b *Block, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	if b.Label() == "" {
		return fmt.Errorf("module block %s has no label", b.FullName())
	}

	v.logger.Debugf("adding module %s to the evaluation context", b.FullName())
	e.ctx.SetByDot(b.Values(), b.LocalName())

	return nil
}

func (v *VertexModuleCall) expand(e *Evaluator, b *Block, mutex *sync.Mutex) ([]*Block, error) {
	mutex.Lock()
	defer mutex.Unlock()
	expanded := []*Block{b}
	expanded = e.expandBlockForEaches(expanded)
	expanded = e.expandBlockCounts(expanded)

	unexpandedName := b.FullName()

	for _, block := range expanded {
		name := block.FullName()

		modCall, err := e.loadModule(block)
		if err != nil {
			return nil, fmt.Errorf("error loading module: %w", err)
		}

		// TODO: do we need this?
		// `evaluateModules` sets the module name to this, so we do it
		// here as well to be consistent
		modCall.Module.Name = modCall.Definition.FullName()

		e.moduleCalls[name] = modCall

		parentContext := NewContext(&hcl.EvalContext{
			Functions: ExpFunctions(modCall.Module.RootPath, e.logger),
		}, nil, e.logger)
		providers := e.getValuesByBlockType("provider")
		for key, provider := range providers.AsValueMap() {
			parentContext.Set(provider, key)
		}

		vars := block.Values().AsValueMap()

		moduleEvaluator := NewEvaluator(
			*modCall.Module,
			e.workingDir,
			vars,
			e.moduleMetadata,
			map[string]map[string]cty.Value{},
			e.workspace,
			e.blockBuilder,
			nil,
			e.logger,
			parentContext,
		)

		v.moduleConfigs.Add(unexpandedName, ModuleConfig{
			name:            name,
			moduleCall:      modCall,
			evaluator:       moduleEvaluator,
			parentEvaluator: e,
		})
	}

	b.expanded = true
	return expanded, nil
}
