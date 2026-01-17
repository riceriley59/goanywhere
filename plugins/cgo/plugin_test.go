package cgo

import (
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
		It("returns cgo", func() {
			Expect(plugin.Name()).To(Equal("cgo"))
		})
	})

	Describe("Generate", func() {
		It("generates valid CGO code for simple package", func() {
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
			Expect(codeStr).To(ContainSubstring("package main"))
			Expect(codeStr).To(ContainSubstring("import \"C\""))
			Expect(codeStr).To(ContainSubstring("//export test_Add"))
			Expect(codeStr).To(ContainSubstring("func main()"))
		})

		It("generates struct wrappers", func() {
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
			Expect(codeStr).To(ContainSubstring("Point_New"))
			Expect(codeStr).To(ContainSubstring("Point_Free"))
			Expect(codeStr).To(ContainSubstring("Point_GetX"))
			Expect(codeStr).To(ContainSubstring("Point_SetX"))
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
			Expect(codeStr).To(ContainSubstring("Point_Distance"))
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
			Expect(codeStr).To(ContainSubstring("test_Greet"))
			Expect(codeStr).To(ContainSubstring("C.GoString"))
			Expect(codeStr).To(ContainSubstring("C.CString"))
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
			Expect(codeStr).To(ContainSubstring("test_Divide"))
			Expect(codeStr).To(ContainSubstring("outError"))
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
			Expect(codeStr).NotTo(ContainSubstring("test_Sum"))
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
			Expect(codeStr).To(ContainSubstring("test_Toggle"))
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
			Expect(codeStr).To(ContainSubstring("test_Multiply"))
		})

		It("handles multiple return values", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "DivMod",
						Params: []core.ParsedParam{
							{Name: "a", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Name: "b", Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
							{Type: core.ParsedType{Kind: core.KindPrimitive, Name: "int"}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("test_DivMod"))
		})

		It("generates handle registry code", func() {
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Structs: []core.ParsedStruct{
					{Name: "Point"},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("handleMap"))
			Expect(codeStr).To(ContainSubstring("registerHandle"))
			Expect(codeStr).To(ContainSubstring("getHandle"))
			Expect(codeStr).To(ContainSubstring("freeHandle"))
		})

		It("generates free functions", func() {
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
			Expect(codeStr).To(ContainSubstring("Free_String"))
			Expect(codeStr).To(ContainSubstring("Free_Bytes"))
		})

		It("handles byte slice parameters", func() {
			elemType := core.ParsedType{Kind: core.KindPrimitive, Name: "byte"}
			pkg := &core.ParsedPackage{
				Name:       "test",
				ImportPath: "github.com/test/test",
				Functions: []core.ParsedFunc{
					{
						Name: "Process",
						Params: []core.ParsedParam{
							{Name: "data", Type: core.ParsedType{Kind: core.KindSlice, Name: "[]byte", ElemType: &elemType}},
						},
						Results: []core.ParsedResult{
							{Type: core.ParsedType{Kind: core.KindSlice, Name: "[]byte", ElemType: &elemType}},
						},
					},
				},
			}

			code, err := plugin.Generate(pkg)
			Expect(err).NotTo(HaveOccurred())

			codeStr := string(code)
			Expect(codeStr).To(ContainSubstring("test_Process"))
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
			Expect(codeStr).To(ContainSubstring("Calculator_Add"))
		})
	})

	Describe("capitalize", func() {
		It("capitalizes first letter", func() {
			Expect(capitalize("hello")).To(Equal("Hello"))
			Expect(capitalize("world")).To(Equal("World"))
		})

		It("handles empty string", func() {
			Expect(capitalize("")).To(Equal(""))
		})

		It("handles already capitalized", func() {
			Expect(capitalize("Hello")).To(Equal("Hello"))
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
			Expect(string(code)).To(ContainSubstring("package main"))
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
			Expect(string(code)).To(ContainSubstring("test_GetValue"))
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
			Expect(string(code)).To(ContainSubstring("test_DoNothing"))
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
			Expect(codeStr).To(ContainSubstring("Mixed_GetPublic"))
			Expect(codeStr).NotTo(ContainSubstring("Mixed_Getprivate"))
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
			Expect(codeStr).To(ContainSubstring("Service_Call"))
			Expect(codeStr).To(ContainSubstring("outError"))
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
			Expect(string(code)).To(ContainSubstring("test_ProcessUint"))
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
			Expect(string(code)).To(ContainSubstring("Value_Get"))
		})
	})

	Describe("getSharedLibExtension", func() {
		It("returns platform-specific extension", func() {
			ext := getSharedLibExtension()
			Expect(ext).To(BeElementOf(".so", ".dylib", ".dll"))
		})
	})
})
