package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexModule struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
}

func (v *VertexModule) ID() string {
	return v.block.FullName()
}

func (v *VertexModule) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexModule) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexModule) References() []VertexReference {
	return referencesForBlock(v.block)
}

func (v *VertexModule) Evaluate() error {
	if v.block.Label() == "" {
		return fmt.Errorf("module block %s has no label", v.ID())
	}

	v.logger.Debugf("adding module %s to the evaluation context", v.ID())
	v.evaluator.ctx.SetByDot(v.block.Values(), v.block.LocalName())

	return nil
}

func (v *VertexModule) Expand() ([]*Block, error) {
	visitMu.Lock()

	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandBlockCounts(expanded)

	moduleEvaluators := map[string]*Evaluator{}

	g, err := NewGraphWithRoot(v.logger)
	if err != nil {
		return nil, fmt.Errorf("error creating new graph: %w", err)
	}

	for _, block := range expanded {
		modCall, err := v.evaluator.loadModule(block)
		if err != nil {
			visitMu.Unlock()
			return nil, fmt.Errorf("error loading module: %w", err)
		}

		// TODO: do we need this?
		// `evaluateModules` sets the module name to this, so we do it
		// here as well to be consistent
		modCall.Module.Name = modCall.Definition.FullName()

		v.evaluator.moduleCalls[block.FullName()] = modCall

		parentContext := NewContext(&hcl.EvalContext{
			Functions: ExpFunctions(modCall.Module.RootPath, v.evaluator.logger),
		}, nil, v.evaluator.logger)
		providers := v.evaluator.getValuesByBlockType("provider")
		for key, provider := range providers.AsValueMap() {
			parentContext.Set(provider, key)
		}

		vars := block.Values().AsValueMap()

		moduleEvaluator := NewEvaluator(
			*modCall.Module,
			v.evaluator.workingDir,
			vars,
			v.evaluator.moduleMetadata,
			map[string]map[string]cty.Value{},
			v.evaluator.workspace,
			v.evaluator.blockBuilder,
			nil,
			v.evaluator.logger,
			parentContext,
		)
		moduleEvaluators[block.FullName()] = moduleEvaluator

		err = g.Populate(moduleEvaluator)
		if err != nil {
			visitMu.Unlock()
			return nil, fmt.Errorf("error populating graph: %w", err)
		}
	}
	visitMu.Unlock()

	g.ReduceTransitively()
	g.Walk()

	visitMu.Lock()

	for fullName, moduleEvaluator := range moduleEvaluators {
		moduleEvaluator.module.Blocks = moduleEvaluator.filteredBlocks

		modCall := v.evaluator.moduleCalls[fullName]
		modCall.Module = moduleEvaluator.collectModules()

		outputs := moduleEvaluator.exportOutputs()
		if val := modCall.Module.Key(); val != nil {
			v.evaluator.ctx.Set(outputs, "module", stripCount(modCall.Name), *val)
		} else if val := modCall.Module.Index(); val != nil {
			v.evaluator.ctx.Set(outputs, "module", stripCount(modCall.Name), fmt.Sprintf("%d", *val))
		} else {
			v.evaluator.ctx.Set(outputs, "module", modCall.Name)
		}
	}

	v.block.expanded = true

	visitMu.Unlock()

	return expanded, nil
}
