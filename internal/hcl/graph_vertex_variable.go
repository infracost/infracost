package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

type VertexVariable struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexVariable) ID() string {
	return v.block.FullName()
}

func (v *VertexVariable) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexVariable) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexVariable) Visit(mutex *sync.Mutex) error {
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

		// Re-evaluate the input variables for the module instance since the inputs
		// to the module might have changed
		inputVars := e.inputVars
		if moduleInstance.moduleCall != nil {
			inputVars = moduleInstance.moduleCall.Definition.values().AsValueMap()
		}

		val, err := e.evaluateVariable(blockInstance, inputVars)
		if err != nil {
			return fmt.Errorf("could not evaluate variable %s: %w", v.ID(), err)
		}

		v.logger.Debug().Msgf("adding variable %s to the evaluation context", v.ID())
		key := fmt.Sprintf("var.%s", blockInstance.Label())
		e.ctx.SetByDot(val, key)

		e.AddFilteredBlocks(blockInstance)
	}

	return nil
}
