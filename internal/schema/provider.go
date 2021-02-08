package schema

type Provider interface {
	LoadResources(map[string]*UsageData) (*State, error)
}
