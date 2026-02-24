package provider

import (
	"fmt"
	"sync"
)

// Registry manages all registered DNS providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]DNSProvider
}

var globalRegistry = &Registry{
	providers: make(map[string]DNSProvider),
}

// Register adds a provider to the global registry
func Register(provider DNSProvider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.providers[provider.Name()] = provider
}

// Get returns a provider by name
func Get(name string) (DNSProvider, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	p, ok := globalRegistry.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}

// List returns info about all registered providers
func List() []ProviderInfo {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	var infos []ProviderInfo
	for _, p := range globalRegistry.providers {
		infos = append(infos, ProviderInfo{
			Name:        p.Name(),
			DisplayName: p.DisplayName(),
		})
	}
	return infos
}
