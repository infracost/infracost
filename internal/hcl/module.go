package hcl

import (
	"fmt"
	"strings"
)

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

	Blocks Blocks
	// RawBlocks are the Blocks that were built when the module was loaded from the filesystem.
	// These are safe to pass to the child module calls as they are yet to be expanded.
	RawBlocks  Blocks
	RootPath   string
	ModulePath string

	Modules  []*Module
	Parent   *Module
	Warnings []Warning
}

// WarningCode is used to delineate warnings across Infracost.
type WarningCode int

const (
	WarningMissingVars WarningCode = iota + 1
)

// Warning holds information about non-critical errors that occurred within a module evaluation.
type Warning struct {
	Code  WarningCode
	Title string
	Data  interface{}

	// FriendlyMessage should be used to display a readable message to the CLI user.
	FriendlyMessage string
}

// NewMissingVarsWarning returns a Warning using the WarningMissingVars error code. It expects that vars
// is a list of Terraform variables that cannot be found in the evaluation context.
func NewMissingVarsWarning(vars []string) Warning {
	return Warning{
		Code:  WarningMissingVars,
		Title: "Missing Terraform vars",
		Data:  vars,
		FriendlyMessage: fmt.Sprintf(
			"Input values were not provided for following Terraform variables: %s. %s",
			joinQuotes(vars),
			"Use --terraform-var-file or --terraform-var to specify them.",
		),
	}
}

func joinQuotes(elems []string) string {
	quoted := make([]string, len(elems))
	for i, elem := range elems {
		quoted[i] = fmt.Sprintf("%q", elem)
	}

	return strings.Join(quoted, ", ")
}
