package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
)

type VertexData struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexData) ID() string {
	return v.block.FullName()
}

func (v *VertexData) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexData) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexData) Visit(mutex *sync.Mutex) error {
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
			return fmt.Errorf("could not find block %q in module %q", v.block.FullName(), moduleInstance.name)
		}

		// Reload the provider references for this module instance
		// We need to do this because we might be evaluating a data block that
		// needs to get data from a provider block, e.g. aws_default_tags.
		// We might be able to improve this by only evaluating the
		// provider block when we need it.
		for name, providerBlock := range e.module.ProviderReferences {
			e.ctx.Set(providerBlock.Values(), name)
		}

		err := v.evaluate(e, blockInstance)
		if err != nil {
			return fmt.Errorf("could not evaluate data block %q", v.ID())
		}

		expanded, err := v.expand(e, blockInstance)
		if err != nil {
			return fmt.Errorf("could not expand data block %q", v.ID())
		}

		e.AddFilteredBlocks(expanded...)
	}

	return nil
}

func (v *VertexData) evaluate(e *Evaluator, b *Block) error {
	if len(b.Labels()) < 2 {
		return fmt.Errorf("data block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := e.ctx.Get(b.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := e.evaluateResourceOrData(b, existingVals)

	v.logger.Debug().Msgf("adding data %s to the evaluation context", v.ID())
	e.ctx.SetByDot(val, fmt.Sprintf("data.%s", b.TypeLabel()))

	return nil
}

func (v *VertexData) expand(e *Evaluator, b *Block) ([]*Block, error) {
	expanded := []*Block{b}
	expanded = e.expandBlockCounts(expanded)
	expanded = e.expandBlockForEaches(expanded)
	expanded = e.expandDynamicBlocks(expanded...)

	b.expanded = true

	return expanded, nil
}
