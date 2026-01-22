// Copyright 2026 Riley Rice
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory

import (
	"fmt"
	"sort"
	"sync"

	"github.com/riceriley59/goanywhere/internal/core"
)

// PluginFactory creates plugins by name
type PluginFactory func(verbose bool) core.Plugin

// registry holds all registered plugin factories
var (
	registryMu sync.RWMutex
	registry   = make(map[string]PluginFactory)
)

// Register adds a plugin factory to the registry
func Register(name string, factory PluginFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if factory == nil {
		panic("factory: Register factory is nil")
	}
	if _, exists := registry[name]; exists {
		panic("factory: Register called twice for plugin " + name)
	}
	registry[name] = factory
}

// Get returns a plugin by name
func Get(name string, verbose bool) (core.Plugin, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown plugin: %s (available: %v)", name, List())
	}
	return factory(verbose), nil
}

// List returns all registered plugin names
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Has checks if a plugin is registered
func Has(name string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := registry[name]
	return ok
}
