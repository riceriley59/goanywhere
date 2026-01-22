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

package integration

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/riceriley59/goanywhere/internal/core"
	"github.com/riceriley59/goanywhere/plugins/cgo"
	"github.com/riceriley59/goanywhere/plugins/python"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = Describe("End-to-end bindings generation", func() {
	var (
		parser     *core.Parser
		fixtureDir string
		pkg        *core.ParsedPackage
	)

	BeforeEach(func() {
		parser = core.NewParser(false)
		wd, _ := os.Getwd()
		fixtureDir = filepath.Join(wd, "..", "fixtures", "simple")

		var err error
		pkg, err = parser.ParsePackage(fixtureDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CGO plugin generation", func() {
		It("generates valid CGO bindings", func() {
			plugin := cgo.NewPlugin(false)
			Expect(plugin.Name()).To(Equal("cgo"))

			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			code := string(output)
			Expect(code).To(ContainSubstring("package main"))
			Expect(code).To(ContainSubstring("import \"C\""))
			Expect(code).To(ContainSubstring("//export"))
		})

		It("generates function wrappers", func() {
			plugin := cgo.NewPlugin(false)
			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			code := string(output)
			Expect(code).To(ContainSubstring("simple_Add"))
			Expect(code).To(ContainSubstring("simple_Greet"))
		})

		It("generates struct wrappers", func() {
			plugin := cgo.NewPlugin(false)
			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			code := string(output)
			Expect(code).To(ContainSubstring("Point_New"))
			Expect(code).To(ContainSubstring("Point_Free"))
		})
	})

	Describe("Python plugin generation", func() {
		It("generates valid Python bindings", func() {
			plugin := python.NewPlugin(false)
			Expect(plugin.Name()).To(Equal("python"))

			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			code := string(output)
			Expect(code).To(ContainSubstring("from ctypes import"))
			Expect(code).To(ContainSubstring("class"))
		})

		It("generates function wrappers", func() {
			plugin := python.NewPlugin(false)
			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			code := string(output)
			Expect(code).To(ContainSubstring("def add"))
			Expect(code).To(ContainSubstring("def greet"))
		})

		It("generates class wrappers for structs", func() {
			plugin := python.NewPlugin(false)
			output, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			code := string(output)
			Expect(code).To(ContainSubstring("class Point"))
		})
	})

	Describe("Parsing variadic functions", func() {
		It("correctly identifies variadic functions", func() {
			var sumFn *core.ParsedFunc
			for i := range pkg.Functions {
				if pkg.Functions[i].Name == "Sum" {
					sumFn = &pkg.Functions[i]
					break
				}
			}
			Expect(sumFn).NotTo(BeNil())
			Expect(sumFn.IsVariadic).To(BeTrue())
		})
	})

	Describe("Parsing functions with errors", func() {
		It("correctly parses error return type", func() {
			var divideFn *core.ParsedFunc
			for i := range pkg.Functions {
				if pkg.Functions[i].Name == "Divide" {
					divideFn = &pkg.Functions[i]
					break
				}
			}
			Expect(divideFn).NotTo(BeNil())
			Expect(divideFn.Results).To(HaveLen(2))
			Expect(divideFn.Results[1].Type.Kind).To(Equal(core.KindError))
		})
	})
})
