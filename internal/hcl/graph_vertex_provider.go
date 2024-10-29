package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
)

type VertexProvider struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexProvider) ID() string {
	alias := v.block.GetAttribute("alias")
	id := v.block.FullName()
	if alias != nil {
		id = fmt.Sprintf("%s.%s", v.block.FullName(), alias.Value().AsString())
	}

	return id
}

func (v *VertexProvider) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexProvider) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexProvider) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	provider := v.block.Label()
	if provider == "" {
		return fmt.Errorf("provider block %s has no label", v.ID())
	}

	moduleInstances := v.moduleConfigs.Get(v.block.ModuleAddress())
	if len(moduleInstances) == 0 {
		return fmt.Errorf("no module instances found for module address %q", v.block.ModuleAddress())
	}

	for _, moduleInstance := range moduleInstances {
		e := moduleInstance.evaluator

		alias := v.block.GetAttribute("alias")

		var blockInstance *Block
		for _, b := range e.module.Blocks {
			if b.LocalName() == v.block.LocalName() && alias.AsString() == b.GetAttribute("alias").AsString() {
				blockInstance = b
				break
			}
		}

		if blockInstance == nil {
			return fmt.Errorf("could not find block %q in module %q", v.ID(), moduleInstance.name)
		}

		// We don't care about the existing values, this is only needed by the legacy evaluator
		val := e.evaluateProvider(blockInstance, map[string]cty.Value{})

		v.logger.Debug().Msgf("adding %s to the evaluation context", v.ID())
		e.ctx.SetByDot(val, blockInstance.Label())

		e.AddFilteredBlocks(blockInstance)
	}

	return nil
}
