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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/riceriley59/goanywhere/internal/core"
)

func TestPythonMapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Python Mapper Suite")
}

var _ = Describe("TypeMapper", func() {
	var mapper *TypeMapper

	BeforeEach(func() {
		structs := []core.ParsedStruct{{Name: "Point"}}
		mapper = NewTypeMapper(structs)
	})

	Describe("MapType primitives", func() {
		DescribeTable("maps primitive types correctly",
			func(goType, expectedCtypes, expectedPy string) {
				pt := core.ParsedType{Kind: core.KindPrimitive, Name: goType}
				pyType, err := mapper.MapType(pt)
				Expect(err).NotTo(HaveOccurred())
				Expect(pyType.CtypesType).To(Equal(expectedCtypes))
				Expect(pyType.PyType).To(Equal(expectedPy))
			},
			Entry("int", "int", "c_longlong", "int"),
			Entry("int8", "int8", "c_int8", "int"),
			Entry("int16", "int16", "c_int16", "int"),
			Entry("int32", "int32", "c_int32", "int"),
			Entry("int64", "int64", "c_int64", "int"),
			Entry("uint", "uint", "c_ulonglong", "int"),
			Entry("uint8", "uint8", "c_uint8", "int"),
			Entry("uint16", "uint16", "c_uint16", "int"),
			Entry("uint32", "uint32", "c_uint32", "int"),
			Entry("uint64", "uint64", "c_uint64", "int"),
			Entry("float32", "float32", "c_float", "float"),
			Entry("float64", "float64", "c_double", "float"),
			Entry("bool", "bool", "c_bool", "bool"),
			Entry("byte", "byte", "c_uint8", "int"),
			Entry("rune", "rune", "c_int32", "int"),
			Entry("uintptr", "uintptr", "c_size_t", "int"),
		)
	})

	Describe("MapType string", func() {
		It("maps string type", func() {
			pt := core.ParsedType{Kind: core.KindString, Name: "string"}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.CtypesType).To(Equal("c_char_p"))
			Expect(pyType.PyType).To(Equal("str"))
			Expect(pyType.IsString).To(BeTrue())
			Expect(pyType.NeedsFree).To(BeTrue())
		})
	})

	Describe("MapType error", func() {
		It("maps error type", func() {
			pt := core.ParsedType{Kind: core.KindError, Name: "error"}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.IsError).To(BeTrue())
		})
	})

	Describe("MapType struct", func() {
		It("maps struct as handle", func() {
			pt := core.ParsedType{Kind: core.KindStruct, Name: "Point"}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.CtypesType).To(Equal("c_size_t"))
			Expect(pyType.IsHandle).To(BeTrue())
		})
	})

	Describe("MapType pointer", func() {
		It("maps pointer to known struct as handle", func() {
			elem := core.ParsedType{Kind: core.KindStruct, Name: "Point"}
			pt := core.ParsedType{Kind: core.KindPointer, Name: "*Point", ElemType: &elem}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.IsHandle).To(BeTrue())
		})

		It("returns error for pointer without element type", func() {
			pt := core.ParsedType{Kind: core.KindPointer, Name: "*int"}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MapType slice", func() {
		It("maps slice", func() {
			elem := core.ParsedType{Kind: core.KindPrimitive, Name: "int"}
			pt := core.ParsedType{Kind: core.KindSlice, Name: "[]int", ElemType: &elem}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.PyType).To(Equal("list"))
		})

		It("returns error for slice without element type", func() {
			pt := core.ParsedType{Kind: core.KindSlice, Name: "[]int"}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MapType array", func() {
		It("maps fixed array", func() {
			elem := core.ParsedType{Kind: core.KindPrimitive, Name: "int"}
			pt := core.ParsedType{Kind: core.KindArray, Name: "[5]int", ElemType: &elem, Size: 5}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.PyType).To(Equal("list"))
		})

		It("returns error for array without element type", func() {
			pt := core.ParsedType{Kind: core.KindArray, Name: "[5]int", Size: 5}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MapType map", func() {
		It("maps map as handle", func() {
			pt := core.ParsedType{Kind: core.KindMap, Name: "map[string]int"}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.PyType).To(Equal("dict"))
			Expect(pyType.IsHandle).To(BeTrue())
		})
	})

	Describe("MapType interface", func() {
		It("maps empty interface", func() {
			pt := core.ParsedType{Kind: core.KindInterface, Name: "interface{}"}
			pyType, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(pyType.IsHandle).To(BeTrue())
		})
	})

	Describe("MapType unsupported", func() {
		It("returns error for chan", func() {
			pt := core.ParsedType{Kind: core.KindChan, Name: "chan int"}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})

		It("returns error for func", func() {
			pt := core.ParsedType{Kind: core.KindFunc, Name: "func()"}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})
	})
})
