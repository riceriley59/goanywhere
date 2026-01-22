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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/riceriley59/goanywhere/internal/core"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Factory Suite")
}

type mockPlugin struct {
	name string
}

func (m *mockPlugin) Name() string                                     { return m.name }
func (m *mockPlugin) Generate(pkg *core.ParsedPackage) ([]byte, error) { return nil, nil }
func (m *mockPlugin) Build(pkg *core.ParsedPackage, inputPath string, opts *core.BuildOptions) error {
	return nil
}

var _ = Describe("Plugin factory", func() {
	Describe("Register and Get", func() {
		It("registers and retrieves a plugin", func() {
			Register("test-plugin", func(verbose bool) core.Plugin {
				return &mockPlugin{name: "test-plugin"}
			})

			plugin, err := Get("test-plugin", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(plugin.Name()).To(Equal("test-plugin"))
		})

		It("returns error for unknown plugin", func() {
			_, err := Get("unknown-plugin", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown plugin"))
		})

		It("panics on nil factory", func() {
			Expect(func() {
				Register("nil-plugin", nil)
			}).To(Panic())
		})

		It("panics on duplicate registration", func() {
			Register("dup-plugin", func(verbose bool) core.Plugin {
				return &mockPlugin{name: "dup-plugin"}
			})

			Expect(func() {
				Register("dup-plugin", func(verbose bool) core.Plugin {
					return &mockPlugin{name: "dup-plugin"}
				})
			}).To(Panic())
		})
	})

	Describe("List", func() {
		It("returns sorted plugin names", func() {
			Register("z-plugin", func(verbose bool) core.Plugin {
				return &mockPlugin{name: "z-plugin"}
			})
			Register("a-plugin", func(verbose bool) core.Plugin {
				return &mockPlugin{name: "a-plugin"}
			})

			list := List()
			Expect(list).To(ContainElement("a-plugin"))
			Expect(list).To(ContainElement("z-plugin"))

			// Verify sorted
			aIdx, zIdx := -1, -1
			for i, name := range list {
				if name == "a-plugin" {
					aIdx = i
				}
				if name == "z-plugin" {
					zIdx = i
				}
			}
			Expect(aIdx).To(BeNumerically("<", zIdx))
		})
	})

	Describe("Has", func() {
		It("returns true for registered plugins", func() {
			Register("exists-plugin", func(verbose bool) core.Plugin {
				return &mockPlugin{name: "exists-plugin"}
			})
			Expect(Has("exists-plugin")).To(BeTrue())
		})

		It("returns false for unregistered plugins", func() {
			Expect(Has("not-exists-plugin")).To(BeFalse())
		})
	})
})
