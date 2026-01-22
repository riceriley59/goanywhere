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
	"fmt"

	"github.com/riceriley59/goanywhere/internal/core"
)

// CType represents a C type with associated metadata
type CType struct {
	CTypeName  string // C type name (e.g., "int64_t", "char*")
	GoTypeName string // Original Go type
	NeedsAlloc bool   // Requires memory allocation on return
	NeedsFree  bool   // Caller must free
	IsHandle   bool   // Use opaque handle pattern
	IsOutParam bool   // Used as out parameter (for errors)
}

// TypeMapper handles Go to C type mapping
type TypeMapper struct {
	structRegistry map[string]*core.ParsedStruct
}

// NewTypeMapper creates a TypeMapper with known structs
func NewTypeMapper(structs []core.ParsedStruct) *TypeMapper {
	registry := make(map[string]*core.ParsedStruct)
	for i := range structs {
		registry[structs[i].Name] = &structs[i]
	}
	return &TypeMapper{
		structRegistry: registry,
	}
}

// MapType converts a ParsedType to CType
func (m *TypeMapper) MapType(pt core.ParsedType) (CType, error) {
	switch pt.Kind {
	case core.KindPrimitive:
		return m.mapPrimitive(pt.Name), nil

	case core.KindString:
		return CType{
			CTypeName:  "*C.char",
			GoTypeName: "string",
			NeedsAlloc: true,
			NeedsFree:  true,
		}, nil

	case core.KindError:
		return CType{
			CTypeName:  "**C.char",
			GoTypeName: "error",
			NeedsAlloc: true,
			NeedsFree:  true,
			IsOutParam: true,
		}, nil

	case core.KindPointer:
		if pt.ElemType == nil {
			return CType{}, fmt.Errorf("pointer type missing element type")
		}
		// Check if it's a pointer to a known struct
		if pt.ElemType.Kind == core.KindStruct {
			if _, ok := m.structRegistry[pt.ElemType.Name]; ok {
				return CType{
					CTypeName:  "C.uintptr_t",
					GoTypeName: "*" + pt.ElemType.Name,
					IsHandle:   true,
				}, nil
			}
		}
		// For other pointer types, try to map the element
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return CType{}, err
		}
		return CType{
			CTypeName:  "*" + elemType.CTypeName,
			GoTypeName: "*" + pt.ElemType.Name,
		}, nil

	case core.KindStruct:
		// Check if it's a known struct in the package
		if _, ok := m.structRegistry[pt.Name]; ok {
			return CType{
				CTypeName:  "C.uintptr_t",
				GoTypeName: pt.Name,
				IsHandle:   true,
			}, nil
		}
		// Unknown struct - use opaque handle
		return CType{
			CTypeName:  "C.uintptr_t",
			GoTypeName: pt.Name,
			IsHandle:   true,
		}, nil

	case core.KindSlice:
		if pt.ElemType == nil {
			return CType{}, fmt.Errorf("slice type missing element type")
		}
		// Slices are represented as pointer + length
		// Special case for []byte
		if pt.ElemType.Kind == core.KindPrimitive && pt.ElemType.Name == "byte" {
			return CType{
				CTypeName:  "unsafe.Pointer", // Will generate data + len params
				GoTypeName: "[]byte",
				NeedsAlloc: true,
				NeedsFree:  true,
			}, nil
		}
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return CType{}, err
		}
		return CType{
			CTypeName:  "*" + elemType.CTypeName, // Pointer to element type
			GoTypeName: pt.Name,
			NeedsAlloc: true,
			NeedsFree:  true,
		}, nil

	case core.KindArray:
		if pt.ElemType == nil {
			return CType{}, fmt.Errorf("array type missing element type")
		}
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return CType{}, err
		}
		return CType{
			CTypeName:  fmt.Sprintf("[%d]%s", pt.Size, elemType.CTypeName),
			GoTypeName: pt.Name,
		}, nil

	case core.KindMap:
		// Maps use opaque handles with accessor functions
		return CType{
			CTypeName:  "C.uintptr_t",
			GoTypeName: pt.Name,
			IsHandle:   true,
		}, nil

	case core.KindInterface:
		// Empty interface uses void*
		return CType{
			CTypeName:  "unsafe.Pointer",
			GoTypeName: "interface{}",
			IsHandle:   true,
		}, nil

	case core.KindChan:
		return CType{}, &core.UnsupportedTypeError{
			Type:   "chan",
			Reason: "channels cannot be exposed via CGO",
		}

	case core.KindFunc:
		return CType{}, &core.UnsupportedTypeError{
			Type:   "func",
			Reason: "function types cannot be exposed via CGO",
		}

	default:
		return CType{}, fmt.Errorf("unknown type kind: %v", pt.Kind)
	}
}

// mapPrimitive maps Go primitive types to C types
func (m *TypeMapper) mapPrimitive(name string) CType {
	switch name {
	case "int":
		return CType{CTypeName: "C.longlong", GoTypeName: "int"}
	case "int8":
		return CType{CTypeName: "C.int8_t", GoTypeName: "int8"}
	case "int16":
		return CType{CTypeName: "C.int16_t", GoTypeName: "int16"}
	case "int32":
		return CType{CTypeName: "C.int32_t", GoTypeName: "int32"}
	case "int64":
		return CType{CTypeName: "C.int64_t", GoTypeName: "int64"}
	case "uint":
		return CType{CTypeName: "C.ulonglong", GoTypeName: "uint"}
	case "uint8", "byte":
		return CType{CTypeName: "C.uint8_t", GoTypeName: "uint8"}
	case "uint16":
		return CType{CTypeName: "C.uint16_t", GoTypeName: "uint16"}
	case "uint32":
		return CType{CTypeName: "C.uint32_t", GoTypeName: "uint32"}
	case "uint64":
		return CType{CTypeName: "C.uint64_t", GoTypeName: "uint64"}
	case "float32":
		return CType{CTypeName: "C.float", GoTypeName: "float32"}
	case "float64":
		return CType{CTypeName: "C.double", GoTypeName: "float64"}
	case "bool":
		return CType{CTypeName: "C.bool", GoTypeName: "bool"}
	case "rune":
		return CType{CTypeName: "C.int32_t", GoTypeName: "rune"}
	case "uintptr":
		return CType{CTypeName: "C.uintptr_t", GoTypeName: "uintptr"}
	default:
		return CType{CTypeName: "C.longlong", GoTypeName: name}
	}
}
