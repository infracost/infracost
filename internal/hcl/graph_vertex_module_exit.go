package hcl

import (
	"sync"

	"github.com/rs/zerolog"
)

type VertexModuleExit struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexModuleExit) ID() string {
	return v.block.FullName()
}

func (v *VertexModuleExit) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexModuleExit) References() []VertexReference {
	return []VertexReference{}
}

func (v *VertexModuleExit) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	moduleInstances := v.moduleConfigs.Get(v.block.FullName())

	for _, moduleInstance := range moduleInstances {
		e := moduleInstance.evaluator
		e.module = *e.collectModules()
		e.module.Blocks = e.filteredBlocks

		modCall := moduleInstance.moduleCall
		if modCall == nil {
			continue
		}
		modCall.Module = &e.module
	}

	return nil
}
