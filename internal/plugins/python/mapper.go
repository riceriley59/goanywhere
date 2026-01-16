package python

import (
	"fmt"

	"github.com/riceriley59/goanywhere/internal/core"
)

// PyType represents a Python ctypes type
type PyType struct {
	CtypesType       string // ctypes type for parameters (e.g., "c_double", "c_char_p")
	CtypesReturnType string // ctypes type for return values (may differ for strings)
	PyType           string // Python type hint (e.g., "float", "str")
	NeedsFree        bool   // Caller must free memory
	IsHandle         bool   // Use handle pattern (for structs)
	IsError          bool   // Error out parameter
	IsString         bool   // Is a string type (needs special handling)
}

// TypeMapper handles Go to Python/ctypes type mapping
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

// MapType converts a ParsedType to PyType
func (m *TypeMapper) MapType(pt core.ParsedType) (PyType, error) {
	switch pt.Kind {
	case core.KindPrimitive:
		return m.mapPrimitive(pt.Name), nil

	case core.KindString:
		return PyType{
			CtypesType:       "c_char_p",  // For input parameters
			CtypesReturnType: "c_void_p",  // For return values (to preserve pointer for freeing)
			PyType:           "str",
			NeedsFree:        true,
			IsString:         true,
		}, nil

	case core.KindError:
		return PyType{
			CtypesType: "POINTER(c_char_p)",
			PyType:     "str",
			IsError:    true,
		}, nil

	case core.KindPointer:
		if pt.ElemType == nil {
			return PyType{}, fmt.Errorf("pointer type missing element type")
		}
		if pt.ElemType.Kind == core.KindStruct {
			if _, ok := m.structRegistry[pt.ElemType.Name]; ok {
				return PyType{
					CtypesType: "c_size_t",
					PyType:     pt.ElemType.Name,
					IsHandle:   true,
				}, nil
			}
		}
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return PyType{}, err
		}
		return PyType{
			CtypesType: "POINTER(" + elemType.CtypesType + ")",
			PyType:     "Any",
		}, nil

	case core.KindStruct:
		return PyType{
			CtypesType: "c_size_t",
			PyType:     pt.Name,
			IsHandle:   true,
		}, nil

	case core.KindSlice:
		if pt.ElemType == nil {
			return PyType{}, fmt.Errorf("slice type missing element type")
		}
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return PyType{}, err
		}
		return PyType{
			CtypesType: "POINTER(" + elemType.CtypesType + ")",
			PyType:     "list",
			NeedsFree:  true,
		}, nil

	case core.KindArray:
		if pt.ElemType == nil {
			return PyType{}, fmt.Errorf("array type missing element type")
		}
		elemType, err := m.MapType(*pt.ElemType)
		if err != nil {
			return PyType{}, err
		}
		return PyType{
			CtypesType: fmt.Sprintf("%s * %d", elemType.CtypesType, pt.Size),
			PyType:     "list",
		}, nil

	case core.KindMap:
		return PyType{
			CtypesType: "c_size_t",
			PyType:     "dict",
			IsHandle:   true,
		}, nil

	case core.KindInterface:
		return PyType{
			CtypesType: "c_void_p",
			PyType:     "Any",
			IsHandle:   true,
		}, nil

	case core.KindChan:
		return PyType{}, &core.UnsupportedTypeError{
			Type:   "chan",
			Reason: "channels cannot be exposed to Python",
		}

	case core.KindFunc:
		return PyType{}, &core.UnsupportedTypeError{
			Type:   "func",
			Reason: "function types cannot be exposed to Python",
		}

	default:
		return PyType{}, fmt.Errorf("unknown type kind: %v", pt.Kind)
	}
}

// mapPrimitive maps Go primitive types to ctypes
func (m *TypeMapper) mapPrimitive(name string) PyType {
	switch name {
	case "int":
		return PyType{CtypesType: "c_longlong", PyType: "int"}
	case "int8":
		return PyType{CtypesType: "c_int8", PyType: "int"}
	case "int16":
		return PyType{CtypesType: "c_int16", PyType: "int"}
	case "int32":
		return PyType{CtypesType: "c_int32", PyType: "int"}
	case "int64":
		return PyType{CtypesType: "c_int64", PyType: "int"}
	case "uint":
		return PyType{CtypesType: "c_ulonglong", PyType: "int"}
	case "uint8", "byte":
		return PyType{CtypesType: "c_uint8", PyType: "int"}
	case "uint16":
		return PyType{CtypesType: "c_uint16", PyType: "int"}
	case "uint32":
		return PyType{CtypesType: "c_uint32", PyType: "int"}
	case "uint64":
		return PyType{CtypesType: "c_uint64", PyType: "int"}
	case "float32":
		return PyType{CtypesType: "c_float", PyType: "float"}
	case "float64":
		return PyType{CtypesType: "c_double", PyType: "float"}
	case "bool":
		return PyType{CtypesType: "c_bool", PyType: "bool"}
	case "rune":
		return PyType{CtypesType: "c_int32", PyType: "int"}
	case "uintptr":
		return PyType{CtypesType: "c_size_t", PyType: "int"}
	default:
		return PyType{CtypesType: "c_longlong", PyType: "int"}
	}
}
