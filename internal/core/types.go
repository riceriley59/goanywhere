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

package core

// TypeKind represents the kind of Go type
type TypeKind int

const (
	KindPrimitive TypeKind = iota
	KindString
	KindStruct
	KindSlice
	KindArray
	KindMap
	KindPointer
	KindInterface
	KindFunc
	KindChan
	KindError
)

// ParsedType represents a Go type with full information
type ParsedType struct {
	Kind        TypeKind
	Name        string      // e.g., "int", "MyStruct"
	PackagePath string      // For imported types
	ElemType    *ParsedType // For slices, arrays, pointers, maps (value type)
	KeyType     *ParsedType // For maps (key type)
	Size        int         // For arrays
	IsPointer   bool
}

// ParsedParam represents a function parameter
type ParsedParam struct {
	Name string
	Type ParsedType
}

// ParsedResult represents a function return value
type ParsedResult struct {
	Name string // May be empty for unnamed returns
	Type ParsedType
}

// ParsedFunc represents an exported Go function
type ParsedFunc struct {
	Name       string
	Doc        string
	Params     []ParsedParam
	Results    []ParsedResult
	IsVariadic bool
}

// ParsedField represents a struct field
type ParsedField struct {
	Name     string
	Type     ParsedType
	Tag      string
	Exported bool
}

// ParsedMethod represents a method on a struct
type ParsedMethod struct {
	Name          string
	Doc           string
	ReceiverName  string
	ReceiverType  string
	ReceiverIsPtr bool
	Params        []ParsedParam
	Results       []ParsedResult
	IsVariadic    bool
}

// ParsedStruct represents a Go struct with its methods
type ParsedStruct struct {
	Name    string
	Doc     string
	Fields  []ParsedField
	Methods []ParsedMethod
}

// ParsedPackage represents a parsed Go package
type ParsedPackage struct {
	Name       string
	ImportPath string
	Dir        string
	Functions  []ParsedFunc
	Structs    []ParsedStruct
}

// UnsupportedTypeError indicates a type that cannot be exported
type UnsupportedTypeError struct {
	Type   string
	Reason string
}

func (e *UnsupportedTypeError) Error() string {
	return "unsupported type " + e.Type + ": " + e.Reason
}
