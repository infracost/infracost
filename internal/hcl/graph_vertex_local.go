package hcl

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type VertexLocal struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
	attr      *Attribute
}

func (v *VertexLocal) ID() string {
	return fmt.Sprintf("%s%s", v.block.FullName(), v.attr.Name())
}

func (v *VertexLocal) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexLocal) References() []VertexReference {
	return referencesForAttribute(v.block, v.attr)
}

func (v *VertexLocal) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	v.logger.Debugf("adding attribute %s to the evaluation context", v.ID())

	key := fmt.Sprintf("local.%s", v.attr.Name())
	v.evaluator.ctx.SetByDot(v.attr.Value(), key)

	return nil
}
