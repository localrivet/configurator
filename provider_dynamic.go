package configurator

// DynamicProvider provides configuration from a dynamic source
type DynamicProvider struct {
	name     string
	loadFunc func(interface{}) error
}

// NewDynamicProvider creates a new dynamic provider
func NewDynamicProvider(name string, loadFunc func(interface{}) error) *DynamicProvider {
	return &DynamicProvider{
		name:     name,
		loadFunc: loadFunc,
	}
}

// Name returns the provider name
func (p *DynamicProvider) Name() string {
	return p.name
}

// Load loads configuration from the dynamic source
func (p *DynamicProvider) Load(cfg interface{}) error {
	return p.loadFunc(cfg)
}

// CustomAPIProvider provides configuration from a custom API
type CustomAPIProvider struct {
	url      string
	apiKey   string
	endpoint string
}

// NewCustomAPIProvider creates a new custom API provider
func NewCustomAPIProvider(url, apiKey, endpoint string) *CustomAPIProvider {
	return &CustomAPIProvider{
		url:      url,
		apiKey:   apiKey,
		endpoint: endpoint,
	}
}

// Name returns the provider name
func (p *CustomAPIProvider) Name() string {
	return "custom"
}

// Load loads configuration from a custom API
func (p *CustomAPIProvider) Load(cfg interface{}) error {
	// Implementation would depend on the specific API
	// This is a placeholder that could be implemented as needed
	return nil
}
