package hcl

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
)

var moduleCallArgs = []string{"source", "version", "for_each", "count", "providers", "depends_on", "lifecycle"}

func moduleCallID(moduleAddress string) string {
	return fmt.Sprintf("call:%s", moduleAddress)
}

func stripModuleCallPrefix(id string) string {
	return strings.TrimPrefix(id, "call:")
}

func attrIsVarInput(name string) bool {
	for _, arg := range moduleCallArgs {
		if name == arg {
			return false
		}
	}

	return true
}

type VertexModuleCall struct {
	logger        zerolog.Logger
	moduleConfigs *ModuleConfigs
	block         *Block
}

func (v *VertexModuleCall) ID() string {
	return moduleCallID(v.block.FullName())
}

func (v *VertexModuleCall) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexModuleCall) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexModuleCall) Visit(mutex *sync.Mutex) error {
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

		err := v.evaluate(e, blockInstance, mutex)
		if err != nil {
			return fmt.Errorf("could not evaluate module %q", v.block.FullName())
		}

		expanded, err := v.expand(e, blockInstance, mutex)
		if err != nil {
			return fmt.Errorf("could not expand module %q", v.block.FullName())
		}

		e.AddFilteredBlocks(expanded...)
	}

	return nil
}

func (v *VertexModuleCall) evaluate(e *Evaluator, b *Block, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	if b.Label() == "" {
		return fmt.Errorf("module block %s has no label", b.FullName())
	}

	v.logger.Debug().Msgf("adding module %s to the evaluation context", b.FullName())
	e.ctx.SetByDot(b.Values(), b.LocalName())

	return nil
}

func (v *VertexModuleCall) expand(e *Evaluator, b *Block, mutex *sync.Mutex) ([]*Block, error) {
	mutex.Lock()
	defer mutex.Unlock()
	expanded := []*Block{b}
	expanded = e.expandBlockForEaches(expanded)
	expanded = e.expandBlockCounts(expanded)

	unexpandedName := v.block.FullName()

	for _, block := range expanded {
		name := block.FullName()

		modCall, err := e.loadModuleWithProviders(block)
		if err != nil {
			return nil, fmt.Errorf("error loading module: %w", err)
		}

		// TODO: do we need this?
		// `evaluateModules` sets the module name to this, so we do it
		// here as well to be consistent
		modCall.Module.Name = modCall.Definition.FullName()

		e.moduleCalls[name] = modCall

		vars := block.Values().AsValueMap()

		moduleEvaluator := NewEvaluator(
			*modCall.Module,
			e.workingDir,
			vars,
			e.moduleMetadata,
			map[string]map[string]cty.Value{},
			e.workspace,
			e.blockBuilder,
			e.logger,
			e.isGraph,
		)

		v.moduleConfigs.Add(unexpandedName, ModuleConfig{
			name:            name,
			moduleCall:      modCall,
			evaluator:       moduleEvaluator,
			parentEvaluator: e,
		})
	}

	b.expanded = true
	return expanded, nil
}
