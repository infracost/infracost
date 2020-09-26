package schema

type RegistryItem struct {
	Name    string
	Aliases []string
	Notes   []string
	RFunc   ResourceFunc
}
