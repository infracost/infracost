package hcl

// ModuleCall represents a call to a defined Module by a parent Module.
type ModuleCall struct {
	// Name the name of the module as specified a the point of definition.
	Name string
	// Path is the path to the local directory containing the HCL for the Module.
	Path string
	// Definition is the actual Block where the ModuleCall happens in a hcl.File
	Definition *Block
	// Module contains the parsed root module that represents this ModuleCall.
	Module *Module
}

// Module encapsulates all the Blocks that are part of a Module in a Terraform project.
type Module struct {
	Name   string
	Source string

	Blocks     Blocks
	RootPath   string
	ModulePath string

	Modules []*Module
	Parent  *Module
}
