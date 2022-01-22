package provider

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alec-rabold/zipspy/pkg/zipspy"
)

// Registry contains an index of registered providers.
type Registry struct {
	providers      map[string]Provider
	providersMutex sync.RWMutex
}

// Provider wraps a zipspy reader to supply provider functionality.
type Provider struct {
	// Protocol identifies the provider from a given location.
	Protocol string
	// CreatePlugin defines how to instantiate a new zipspy plugin.
	CreatePlugin func(location string) (zipspy.Reader, error)
}

type registryOption func(*Registry)

// NewRegistry instantiates a new providers registry.
func NewRegistry(opts ...registryOption) *Registry {
	r := &Registry{
		providers: make(map[string]Provider),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithProvider is a helper function to register plugins during registry creation.
func WithProvider(name, protocol string, createPlugin func(location string) (zipspy.Reader, error)) registryOption {
	return func(r *Registry) {
		r.MustRegisterProvider(name, protocol, createPlugin)
	}
}

// RegisterProvider defines a new provider with the given name and protocol.
func (r *Registry) RegisterProvider(name, protocol string, createPlugin func(location string) (zipspy.Reader, error)) error {
	r.providersMutex.Lock()
	defer r.providersMutex.Unlock()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("plugin with name %s already exists", name)
	}
	for name, provider := range r.providers {
		if protocol == provider.Protocol {
			return fmt.Errorf("plugin with name %s already implements protocol %s", name, protocol)
		}
	}
	r.providers[name] = Provider{
		Protocol:     protocol,
		CreatePlugin: createPlugin,
	}
	return nil
}

// MustRegisterProvider is a helper function that panics when RegisterProvider fails.
func (r *Registry) MustRegisterProvider(name, protocol string, createPlugin func(location string) (zipspy.Reader, error)) {
	if err := r.RegisterProvider(name, protocol, createPlugin); err != nil {
		panic(err)
	}
}

// GetProvider returns a new provider for the given location.
func (r *Registry) GetPlugin(location string) (zipspy.Reader, error) {
	r.providersMutex.RLock()
	defer r.providersMutex.RUnlock()
	for _, provider := range r.providers {
		if strings.HasPrefix(location, provider.Protocol) {
			return provider.CreatePlugin(strings.TrimPrefix(location, provider.Protocol))
		}
	}
	return nil, fmt.Errorf("unsupported provider for location %s", location)
}
