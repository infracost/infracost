package hcl

import (
	"fmt"
	"sync"

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

func (v *VertexModule) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexModule) Visit(mutex *sync.Mutex) error {
	err := v.evaluate(mutex)
	if err != nil {
		return fmt.Errorf("could not evaluate module %q", v.ID())
	}

	expanded, err := v.expand(mutex)
	if err != nil {
		return fmt.Errorf("could not expand module %q", v.ID())
	}

	v.evaluator.AddFilteredBlocks(expanded...)
	return nil
}

func (v *VertexModule) evaluate(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	if v.block.Label() == "" {
		return fmt.Errorf("module block %s has no label", v.ID())
	}

	v.logger.Debugf("adding module %s to the evaluation context", v.ID())
	v.evaluator.ctx.SetByDot(v.block.Values(), v.block.LocalName())

	return nil
}

func (v *VertexModule) expand(mutex *sync.Mutex) ([]*Block, error) {
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Printf("expanding module %q\n", v.block.FullName())
	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandBlockCounts(expanded)

	for _, block := range expanded {
		name := block.FullName()
		fmt.Printf("found expanded module %q\n", name)
		modCall, err := v.evaluator.loadModule(block)
		if err != nil {
			return nil, fmt.Errorf("error loading module: %w", err)
		}

		// TODO: do we need this?
		// `evaluateModules` sets the module name to this, so we do it
		// here as well to be consistent
		modCall.Module.Name = modCall.Definition.FullName()

		v.evaluator.moduleCalls[name] = modCall

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
		g, err := NewGraphWithRoot(v.logger, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating new graph: %w", err)
		}

		v.logger.Debugf("evaluating graph for child module %q", name)
		modCall.Module, err = g.Run(moduleEvaluator)
		if err != nil {
			return nil, fmt.Errorf("error populating graph: %w", err)
		}

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
	return expanded, nil
}
