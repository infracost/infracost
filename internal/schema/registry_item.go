package schema

type RegistryItem struct {
	Name    string
	Notes   []string
	RFunc   ResourceFunc
	NoPrice bool
}
