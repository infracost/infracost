package hcl

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

type VertexLocal struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
	attr          *Attribute
}

func (v *VertexLocal) ID() string {
	return fmt.Sprintf("%s%s", v.block.FullName(), v.attr.Name())
}

func (v *VertexLocal) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexLocal) References() []VertexReference {
	return v.attr.VerticesReferenced(v.block)
}

func (v *VertexLocal) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	moduleInstances := v.moduleConfigs.Get(v.block.ModuleAddress())
	if len(moduleInstances) == 0 {
		return fmt.Errorf("no module instances found for module address %q", v.block.ModuleAddress())
	}

	for _, moduleInstance := range moduleInstances {
		e := moduleInstance.evaluator

		v.logger.Debug().Msgf("adding attribute %s to the evaluation context", v.ID())

		var attrInstance *Attribute
		for _, b := range e.module.Blocks {
			if b.LocalName() == v.block.LocalName() {
				attrInstance = b.GetAttribute(v.attr.Name())
				if attrInstance != nil {
					break
				}
			}
		}

		if attrInstance == nil {
			return fmt.Errorf("could not find attribute %q in module %q", v.ID(), moduleInstance.name)
		}

		key := fmt.Sprintf("local.%s", attrInstance.Name())
		e.ctx.SetByDot(attrInstance.Value(), key)
	}

	return nil
}
