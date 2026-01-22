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

package complex

// Config represents a configuration with various types
type Config struct {
	Name    string
	Values  []int
	Data    [5]byte
	Options map[string]string
}

// ProcessArray takes a fixed-size array
func ProcessArray(data [10]int) [10]int {
	return data
}

// ProcessSlice takes a slice
func ProcessSlice(data []string) []string {
	return data
}

// ProcessMap takes a map
func ProcessMap(data map[string]int) map[string]int {
	return data
}

// ProcessPointer takes a pointer
func ProcessPointer(p *Config) *Config {
	return p
}

// ProcessInterface takes an interface
func ProcessInterface(data interface{}) interface{} {
	return data
}

// NewConfig creates a new Config
func NewConfig(name string) *Config {
	return &Config{Name: name}
}

// GetName returns the config name
func (c *Config) GetName() string {
	return c.Name
}

// SetValues sets the values slice
func (c *Config) SetValues(values []int) {
	c.Values = values
}
