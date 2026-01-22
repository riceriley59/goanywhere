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

package cgo

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/riceriley59/goanywhere/internal/core"
)

func TestCgoMapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CGO Mapper Suite")
}

var _ = Describe("TypeMapper", func() {
	var mapper *TypeMapper

	BeforeEach(func() {
		structs := []core.ParsedStruct{{Name: "Point"}}
		mapper = NewTypeMapper(structs)
	})

	Describe("MapType primitives", func() {
		DescribeTable("maps primitive types correctly",
			func(goType, expectedC string) {
				pt := core.ParsedType{Kind: core.KindPrimitive, Name: goType}
				ct, err := mapper.MapType(pt)
				Expect(err).NotTo(HaveOccurred())
				Expect(ct.CTypeName).To(Equal(expectedC))
			},
			Entry("int", "int", "C.longlong"),
			Entry("int8", "int8", "C.int8_t"),
			Entry("int16", "int16", "C.int16_t"),
			Entry("int32", "int32", "C.int32_t"),
			Entry("int64", "int64", "C.int64_t"),
			Entry("uint", "uint", "C.ulonglong"),
			Entry("uint8", "uint8", "C.uint8_t"),
			Entry("uint16", "uint16", "C.uint16_t"),
			Entry("uint32", "uint32", "C.uint32_t"),
			Entry("uint64", "uint64", "C.uint64_t"),
			Entry("float32", "float32", "C.float"),
			Entry("float64", "float64", "C.double"),
			Entry("bool", "bool", "C.bool"),
			Entry("byte", "byte", "C.uint8_t"),
			Entry("rune", "rune", "C.int32_t"),
			Entry("uintptr", "uintptr", "C.uintptr_t"),
		)
	})

	Describe("MapType string", func() {
		It("maps string type", func() {
			pt := core.ParsedType{Kind: core.KindString, Name: "string"}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.CTypeName).To(Equal("*C.char"))
			Expect(ct.NeedsAlloc).To(BeTrue())
			Expect(ct.NeedsFree).To(BeTrue())
		})
	})

	Describe("MapType error", func() {
		It("maps error type", func() {
			pt := core.ParsedType{Kind: core.KindError, Name: "error"}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.CTypeName).To(Equal("**C.char"))
			Expect(ct.IsOutParam).To(BeTrue())
		})
	})

	Describe("MapType struct", func() {
		It("maps known struct as handle", func() {
			pt := core.ParsedType{Kind: core.KindStruct, Name: "Point"}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.CTypeName).To(Equal("C.uintptr_t"))
			Expect(ct.IsHandle).To(BeTrue())
		})

		It("maps unknown struct as handle", func() {
			pt := core.ParsedType{Kind: core.KindStruct, Name: "Unknown"}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.IsHandle).To(BeTrue())
		})
	})

	Describe("MapType pointer", func() {
		It("maps pointer to known struct as handle", func() {
			elem := core.ParsedType{Kind: core.KindStruct, Name: "Point"}
			pt := core.ParsedType{Kind: core.KindPointer, Name: "*Point", ElemType: &elem}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.IsHandle).To(BeTrue())
		})

		It("returns error for pointer without element type", func() {
			pt := core.ParsedType{Kind: core.KindPointer, Name: "*int"}
			_, err := mapper.MapType(pt)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MapType slice", func() {
		It("maps byte slice", func() {
			elem := core.ParsedType{Kind: core.KindPrimitive, Name: "byte"}
			pt := core.ParsedType{Kind: core.KindSlice, Name: "[]byte", ElemType: &elem}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.GoTypeName).To(Equal("[]byte"))
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
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.CTypeName).To(ContainSubstring("[5]"))
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
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.IsHandle).To(BeTrue())
		})
	})

	Describe("MapType interface", func() {
		It("maps empty interface", func() {
			pt := core.ParsedType{Kind: core.KindInterface, Name: "interface{}"}
			ct, err := mapper.MapType(pt)
			Expect(err).NotTo(HaveOccurred())
			Expect(ct.IsHandle).To(BeTrue())
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
