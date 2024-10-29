package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
)

type VertexResource struct {
	moduleConfigs *ModuleConfigs
	logger        zerolog.Logger
	block         *Block
}

func (v *VertexResource) ID() string {
	return v.block.FullName()
}

func (v *VertexResource) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexResource) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexResource) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

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

		err := v.evaluate(e, blockInstance)
		if err != nil {
			return fmt.Errorf("could not evaluate resource block %q", v.ID())
		}

		expanded, err := v.expand(e, blockInstance)
		if err != nil {
			return fmt.Errorf("could not expand resource block %q", v.ID())
		}

		e.AddFilteredBlocks(expanded...)
	}

	return nil
}

func (v *VertexResource) evaluate(e *Evaluator, b *Block) error {
	if len(b.Labels()) < 2 {
		return fmt.Errorf("resource block %s has no label", v.ID())
	}

	// We don't care about the existing values, this is only needed by the legacy evaluator
	val := e.evaluateResourceOrData(b, map[string]cty.Value{})

	v.logger.Debug().Msgf("adding resource %s to the evaluation context", v.ID())
	e.ctx.SetByDot(val, b.TypeLabel())

	return nil
}

func (v *VertexResource) expand(e *Evaluator, b *Block) ([]*Block, error) {
	expanded := []*Block{b}
	expanded = e.expandBlockCounts(expanded)
	expanded = e.expandBlockForEaches(expanded)
	expanded = e.expandDynamicBlocks(expanded...)

	b.expanded = true

	return expanded, nil
}
