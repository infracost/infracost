package schema

type Provider interface {
	Type() string
	DisplayType() string
	LoadResources(map[string]*UsageData) (*Project, error)
}
