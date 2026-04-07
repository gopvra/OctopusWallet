package chain

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu     sync.RWMutex
	chains map[string]Chain
}

func NewRegistry() *Registry {
	return &Registry{
		chains: make(map[string]Chain),
	}
}

func (r *Registry) Register(c Chain) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chains[c.Name()] = c
}

func (r *Registry) Get(name string) (Chain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.chains[name]
	if !ok {
		return nil, fmt.Errorf("chain %q not registered", name)
	}
	return c, nil
}

func (r *Registry) All() []Chain {
	r.mu.RLock()
	defer r.mu.RUnlock()
	chains := make([]Chain, 0, len(r.chains))
	for _, c := range r.chains {
		chains = append(chains, c)
	}
	return chains
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.chains))
	for name := range r.chains {
		names = append(names, name)
	}
	return names
}
