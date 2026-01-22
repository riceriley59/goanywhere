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

package python

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/riceriley59/goanywhere/internal/core"
)

var _ = Describe("Plugin", func() {
	var plugin *Plugin

	BeforeEach(func() {
		plugin = NewPlugin(false)
	})

	Describe("NewPlugin", func() {
		It("creates plugin with verbose off", func() {
			p := NewPlugin(false)
			Expect(p).NotTo(BeNil())
			Expect(p.verbose).To(BeFalse())
		})

		It("creates plugin with verbose on", func() {
			p := NewPlugin(true)
			Expect(p.verbose).To(BeTrue())
		})
	})

	Describe("Name", func() {
		It("returns python", func() {
			Expect(plugin.Name()).To(Equal("python"))
		})
	})

	Describe("Generate", func() {
		It("generates valid Python code for simple package", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Add",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).NotTo(BeEmpty())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("from ctypes import"))
			Expect(codeStr).To(ContainSubstring("def add("))
			Expect(codeStr).To(ContainSubstring("test_Add"))
		})

		It("generates class wrappers for structs", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Point",
						Fields: []core.ParsedField{
							{Name: "X", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}, Exported: true},
							{Name: "Y", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}, Exported: true},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("class Point"))
			Expect(codeStr).To(ContainSubstring("def __init__"))
			Expect(codeStr).To(ContainSubstring("def __del__"))
			Expect(codeStr).To(ContainSubstring("@property"))
			Expect(codeStr).To(ContainSubstring("def x(self)"))
		})

		It("generates method wrappers", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Point",
						Methods: []core.ParsedMethod{
							{
								Name:          "Distance",
								ReceiverName:  "p",
								ReceiverType:  "Point",
								ReceiverIsPtr: true,
								Results: []core.ParsedResult{
									{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "float64"}},
								},
							},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def distance(self)"))
		})

		It("handles string parameters and returns", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Greet",
						Params: []core.ParsedParam{
							{Name: "name", Type: core.ParsedType{Kind: core.KindString, Name: "string"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindString, Name: "string"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def greet("))
			Expect(codeStr).To(ContainSubstring("_encode_string"))
			Expect(codeStr).To(ContainSubstring("_decode_string"))
		})

		It("handles error returns", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Divide",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Type: core.ParsedType{Kind: core.KindError, Name: "error"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def divide("))
			Expect(codeStr).To(ContainSubstring("_check_error"))
		})

		It("skips variadic functions", func() {
			verbosePlugin := NewPlugin(true)
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name:       "Sum",
						IsVariadic: true,
						Params: []core.ParsedParam{
							{Name: "nums", Type: core.ParsedType{Kind: core.KindSlice, Name: "[]int"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
					},
				},
			}

			code, err := verbosePlugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).NotTo(ContainSubstring("def sum("))
		})

		It("generates library loader code", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Hello",
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindString, Name: "string"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def load_library"))
			Expect(codeStr).To(ContainSubstring("libtest"))
			Expect(codeStr).To(ContainSubstring(".so"))
			Expect(codeStr).To(ContainSubstring(".dylib"))
			Expect(codeStr).To(ContainSubstring(".dll"))
		})

		It("generates helper functions", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "GetName",
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindString, Name: "string"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def _encode_string"))
			Expect(codeStr).To(ContainSubstring("def _decode_string"))
			Expect(codeStr).To(ContainSubstring("def _check_error"))
		})

		It("generates context manager support", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{Name: "Resource"},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def __enter__"))
			Expect(codeStr).To(ContainSubstring("def __exit__"))
		})

		It("handles bool parameters", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Toggle",
						Params: []core.ParsedParam{
							{Name: "flag", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "bool"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "bool"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def toggle("))
			Expect(codeStr).To(ContainSubstring("bool"))
		})

		It("handles float parameters", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Multiply",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "float64"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "float64"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "float64"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def multiply("))
			Expect(codeStr).To(ContainSubstring("float"))
		})

		It("handles methods with parameters", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Calculator",
						Methods: []core.ParsedMethod{
							{
								Name:          "Add",
								ReceiverName:  "c",
								ReceiverType:  "Calculator",
								ReceiverIsPtr: true,
								Params: []core.ParsedParam{
									{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
									{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
								},
								Results: []core.ParsedResult{
									{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
								},
							},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def add(self"))
		})

		It("generates type hints", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Add",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("-> int"))
		})
	})

	Describe("toSnakeCase", func() {
		It("converts camelCase to snake_case", func() {
			Expect(toSnakeCase("helloWorld")).To(Equal("hello_world"))
			Expect(toSnakeCase("MyFunction")).To(Equal("my_function"))
			Expect(toSnakeCase("HTTPServer")).To(Equal("h_t_t_p_server"))
		})

		It("handles single word", func() {
			Expect(toSnakeCase("hello")).To(Equal("hello"))
			Expect(toSnakeCase("Hello")).To(Equal("hello"))
		})

		It("handles empty string", func() {
			Expect(toSnakeCase("")).To(Equal(""))
		})
	})

	Describe("Generate comprehensive", func() {
		It("handles empty package", func() {
			pkg := &core.ParsedPackage{
				Name:       "empty",
				ImportPath: "github.com/test/empty",
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("from ctypes import"))
		})

		It("handles function with no params", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name:    "GetValue",
						Results: []core.ParsedResult{{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}}},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("def get_value()"))
		})

		It("handles function with no return value", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "DoNothing",
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("def do_nothing()"))
		})

		It("handles struct with unexported fields", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Mixed",
						Fields: []core.ParsedField{
							{Name: "Public", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}, Exported: true},
							{Name: "private", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}, Exported: false},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("class Mixed"))
			Expect(codeStr).To(ContainSubstring("def public"))
		})

		It("handles method with error return", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Service",
						Methods: []core.ParsedMethod{
							{
								Name:          "Call",
								ReceiverName:  "s",
								ReceiverType:  "Service",
								ReceiverIsPtr: true,
								Results: []core.ParsedResult{
									{Type: core.ParsedType{Kind: core.KindString, Name: "string"}},
									{Type: core.ParsedType{Kind: core.KindError, Name: "error"}},
								},
							},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("class Service"))
			Expect(codeStr).To(ContainSubstring("def call"))
		})

		It("handles uint types", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "ProcessUint",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "uint"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "uint64"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "uint32"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("def process_uint"))
		})

		It("handles value receiver method", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Value",
						Methods: []core.ParsedMethod{
							{
								Name:          "Get",
								ReceiverName:  "v",
								ReceiverType:  "Value",
								ReceiverIsPtr: false,
								Results: []core.ParsedResult{
									{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
								},
							},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("def get(self)"))
		})

		It("handles multiple functions", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{Name: "First", Results: []core.ParsedResult{{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}}}},
					{Name: "Second", Results: []core.ParsedResult{{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}}}},
					{Name: "Third", Results: []core.ParsedResult{{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}}}},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("def first"))
			Expect(codeStr).To(ContainSubstring("def second"))
			Expect(codeStr).To(ContainSubstring("def third"))
		})

		It("handles struct with method that returns struct", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{
						Name: "Builder",
						Methods: []core.ParsedMethod{
							{
								Name:          "Build",
								ReceiverName:  "b",
								ReceiverType:  "Builder",
								ReceiverIsPtr: true,
								Results: []core.ParsedResult{
									{Type: core.ParsedType{Kind: core.KindPointer, Name: "*Builder", ElemType: &core.ParsedType{Kind: core.KindStruct, Name: "Builder"}}},
								},
							},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(code)).To(ContainSubstring("def build"))
		})
	})

	Describe("getSharedLibExtension", func() {
		It("returns platform-specific extension", func() {
			ext := getSharedLibExtension()
			Expect(ext).To(BeElementOf(".so", ".dylib", ".dll"))
		})
	})

	Describe("generatePyprojectToml", func() {
		It("generates setuptools config", func() {
			content := generatePyprojectToml("mypackage", "setuptools")
			Expect(content).To(ContainSubstring("setuptools"))
			Expect(content).To(ContainSubstring("mypackage"))
		})

		It("generates hatch config", func() {
			content := generatePyprojectToml("mypackage", "hatch")
			Expect(content).To(ContainSubstring("hatchling"))
			Expect(content).To(ContainSubstring("mypackage"))
		})

		It("generates poetry config", func() {
			content := generatePyprojectToml("mypackage", "poetry")
			Expect(content).To(ContainSubstring("poetry"))
			Expect(content).To(ContainSubstring("mypackage"))
		})

		It("generates uv config", func() {
			content := generatePyprojectToml("mypackage", "uv")
			Expect(content).To(ContainSubstring("hatchling"))
			Expect(content).To(ContainSubstring("mypackage"))
		})

		It("defaults to setuptools for unknown system", func() {
			content := generatePyprojectToml("mypackage", "unknown")
			Expect(content).To(ContainSubstring("setuptools"))
		})
	})

	Describe("copyFile", func() {
		It("copies file contents", func() {
			// Create source file
			srcFile, err := os.CreateTemp("", "src")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove(srcFile.Name()) }()

			content := []byte("test content")
			_, err = srcFile.Write(content)
			Expect(err).NotTo(HaveOccurred())
			_ = srcFile.Close()

			// Copy to destination
			dstFile, err := os.CreateTemp("", "dst")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove(dstFile.Name()) }()
			_ = dstFile.Close()

			err = copyFile(srcFile.Name(), dstFile.Name())
			Expect(err).NotTo(HaveOccurred())

			// Verify content
			data, err := os.ReadFile(dstFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(content))
		})

		It("returns error for non-existent source", func() {
			err := copyFile("/nonexistent/file", "/tmp/dst")
			Expect(err).To(HaveOccurred())
		})
	})
})
