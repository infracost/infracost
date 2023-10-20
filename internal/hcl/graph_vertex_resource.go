package hcl

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexResource struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
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

	err := v.evaluate()
	if err != nil {
		return fmt.Errorf("could not evaluate resource block %q", v.ID())
	}

	expanded, err := v.expand()
	if err != nil {
		return fmt.Errorf("could not expand resource block %q", v.ID())
	}

	v.evaluator.AddFilteredBlocks(expanded...)
	return nil
}

func (v *VertexResource) evaluate() error {
	if len(v.block.Labels()) < 2 {
		return fmt.Errorf("resource block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := v.evaluator.ctx.Get(v.block.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := v.evaluator.evaluateResource(v.block, existingVals)

	v.logger.Debugf("adding resource %s to the evaluation context", v.ID())
	v.evaluator.ctx.SetByDot(val, v.block.TypeLabel())

	return nil
}

func (v *VertexResource) expand() ([]*Block, error) {
	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockCounts(expanded)
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandDynamicBlocks(expanded...)

	v.block.expanded = true

	return expanded, nil
}
