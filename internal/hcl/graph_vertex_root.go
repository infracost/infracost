package hcl

import "sync"

type VertexRoot struct{}

func (v *VertexRoot) ID() string {
	return "_root"
}

func (v *VertexRoot) ModuleAddress() string {
	return ""
}

func (v *VertexRoot) References() []VertexReference {
	return []VertexReference{}
}

func (v *VertexRoot) Visit(mutex *sync.Mutex) error {
	return nil
}
