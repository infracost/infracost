package block

import "fmt"

type Type struct {
	name                  string
	refName               string
	removeTypeInReference bool
}

func (t Type) Name() string {
	return t.name
}

func (t Type) ShortName() string {
	if t.refName != "" {
		return t.refName
	}
	return t.name
}

var TypeData = Type{
	name: "data",
}

var TypeResource = Type{
	name:                  "resource",
	removeTypeInReference: true,
}

var TypeVariable = Type{
	name:    "variable",
	refName: "var",
}

var TypeLocal = Type{
	name:    "locals",
	refName: "local",
}

var TypeProvider = Type{
	name: "provider",
}

var TypeOutput = Type{
	name: "output",
}

var TypeModule = Type{
	name: "module",
}

var TypeTerraform = Type{
	name: "terraform",
}

var ValidTypes = []Type{
	TypeData,
	TypeLocal,
	TypeModule,
	TypeOutput,
	TypeProvider,
	TypeResource,
	TypeTerraform,
	TypeVariable,
}

func TypeFromRefName(name string) (*Type, error) {
	for _, valid := range ValidTypes {
		if valid.refName == name || (valid.refName == "" && valid.name == name) {
			return &valid, nil
		}
	}
	return nil, fmt.Errorf("block type not found")
}
