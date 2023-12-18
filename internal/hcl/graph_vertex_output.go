package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

type VertexOutput struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexOutput) ID() string {
	if v.block.ModuleName() == "" {
		return v.block.Label()
	}

	return fmt.Sprintf("%s.%s", v.block.ModuleAddress(), v.block.Label())
}

func (v *VertexOutput) ModuleAddress() string {
	return v.block.ModuleAddress()
}
func (v *VertexOutput) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexOutput) Visit(mutex *sync.Mutex) error {
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

		key := fmt.Sprintf("output.%s", blockInstance.Label())
		val, err := e.evaluateOutput(blockInstance)
		if err != nil {
			return fmt.Errorf("could not evaluate output %s: %w", v.ID(), err)
		}

		v.logger.Debug().Msgf("adding output %s to the evaluation context", v.ID())
		e.ctx.SetByDot(val, key)

		parentEvaluator := moduleInstance.parentEvaluator
		modCall := moduleInstance.moduleCall
		if parentEvaluator != nil && modCall != nil {

			var parentKeyParts []string
			if v := modCall.Module.Key(); v != nil {
				parentKeyParts = []string{"module", stripCount(modCall.Name), *v}
			} else if v := modCall.Module.Index(); v != nil {
				parentKeyParts = []string{"module", stripCount(modCall.Name), fmt.Sprintf("%d", *v)}
			} else {
				parentKeyParts = []string{"module", modCall.Name}
			}

			parentKeyParts = append(parentKeyParts, blockInstance.Label())

			parentEvaluator.ctx.Set(val, parentKeyParts...)
		}

		e.AddFilteredBlocks(blockInstance)
	}

	return nil
}
