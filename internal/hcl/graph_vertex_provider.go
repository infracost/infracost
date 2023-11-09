package hcl

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexProvider struct {
	logger        *logrus.Entry
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
			if b.LocalName() == v.block.LocalName() && alias == b.GetAttribute("alias") {
				blockInstance = b
				break
			}
		}

		if blockInstance == nil {
			return fmt.Errorf("could not find block %q in module %q", v.ID(), moduleInstance.name)
		}

		var existingVals map[string]cty.Value
		existingCtx := e.ctx.Get(blockInstance.TypeLabel())
		if !existingCtx.IsNull() {
			existingVals = existingCtx.AsValueMap()
		} else {
			existingVals = make(map[string]cty.Value)
		}

		val := e.evaluateProvider(blockInstance, existingVals)

		v.logger.Debugf("adding %s to the evaluation context", v.ID())
		e.ctx.Set(val, blockInstance.Label())

		e.AddFilteredBlocks(blockInstance)
	}

	return nil
}
