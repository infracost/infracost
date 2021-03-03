package schema

type Provider interface {
	Type() string
	LoadResources(map[string]*UsageData) (*Project, error)
}
