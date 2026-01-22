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

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Parser handles Go source file parsing
type Parser struct {
	fset    *token.FileSet
	verbose bool
}

// NewParser creates a new Parser instance
func NewParser(verbose bool) *Parser {
	return &Parser{
		fset:    token.NewFileSet(),
		verbose: verbose,
	}
}

// ParsePackage parses all .go files in a directory
func (p *Parser) ParsePackage(dirPath string) (*ParsedPackage, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	pkgs, err := parser.ParseDir(p.fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages found in %s", absPath)
	}

	// Take the first non-test package
	var pkg *ast.Package //nolint:staticcheck // ast.Package is returned by parser.ParseDir
	var pkgName string
	for name, p := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			pkg = p
			pkgName = name
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("no non-test packages found in %s", absPath)
	}

	parsed := &ParsedPackage{
		Name: pkgName,
		Dir:  absPath,
	}

	// Collect all methods first to associate with structs later
	methodsByReceiver := make(map[string][]ParsedMethod)

	// Parse all files in the package
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil {
					// This is a method
					method, receiverType, err := p.parseMethod(d)
					if err != nil {
						if p.verbose {
							fmt.Printf("Skipping method %s: %v\n", d.Name.Name, err)
						}
						continue
					}
					if method != nil {
						methodsByReceiver[receiverType] = append(methodsByReceiver[receiverType], *method)
					}
				} else {
					// This is a function
					fn, err := p.parseFunc(d)
					if err != nil {
						if p.verbose {
							fmt.Printf("Skipping function %s: %v\n", d.Name.Name, err)
						}
						continue
					}
					if fn != nil {
						parsed.Functions = append(parsed.Functions, *fn)
					}
				}
			case *ast.GenDecl:
				if d.Tok == token.TYPE {
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						st, ok := ts.Type.(*ast.StructType)
						if !ok {
							continue
						}
						parsedStruct, err := p.parseStruct(ts, st, d.Doc)
						if err != nil {
							if p.verbose {
								fmt.Printf("Skipping struct %s: %v\n", ts.Name.Name, err)
							}
							continue
						}
						if parsedStruct != nil {
							parsed.Structs = append(parsed.Structs, *parsedStruct)
						}
					}
				}
			}
		}
	}

	// Associate methods with structs
	for i := range parsed.Structs {
		structName := parsed.Structs[i].Name
		if methods, ok := methodsByReceiver[structName]; ok {
			parsed.Structs[i].Methods = methods
		}
	}

	return parsed, nil
}

// parseFunc extracts function information from ast.FuncDecl
func (p *Parser) parseFunc(fn *ast.FuncDecl) (*ParsedFunc, error) {
	// Skip unexported functions
	if !isExported(fn.Name.Name) {
		return nil, nil
	}

	parsed := &ParsedFunc{
		Name: fn.Name.Name,
	}

	if fn.Doc != nil {
		parsed.Doc = fn.Doc.Text()
	}

	// Parse parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			pt, err := p.parseType(field.Type)
			if err != nil {
				return nil, err
			}

			// Check for variadic
			if _, isEllipsis := field.Type.(*ast.Ellipsis); isEllipsis {
				parsed.IsVariadic = true
			}

			if len(field.Names) == 0 {
				// Unnamed parameter
				parsed.Params = append(parsed.Params, ParsedParam{Type: pt})
			} else {
				for _, name := range field.Names {
					parsed.Params = append(parsed.Params, ParsedParam{
						Name: name.Name,
						Type: pt,
					})
				}
			}
		}
	}

	// Parse results
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			pt, err := p.parseType(field.Type)
			if err != nil {
				return nil, err
			}

			if len(field.Names) == 0 {
				parsed.Results = append(parsed.Results, ParsedResult{Type: pt})
			} else {
				for _, name := range field.Names {
					parsed.Results = append(parsed.Results, ParsedResult{
						Name: name.Name,
						Type: pt,
					})
				}
			}
		}
	}

	return parsed, nil
}

// parseMethod extracts method information from ast.FuncDecl
func (p *Parser) parseMethod(fn *ast.FuncDecl) (*ParsedMethod, string, error) {
	// Skip unexported methods
	if !isExported(fn.Name.Name) {
		return nil, "", nil
	}

	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return nil, "", nil
	}

	recv := fn.Recv.List[0]
	var receiverName, receiverType string
	var receiverIsPtr bool

	// Get receiver name
	if len(recv.Names) > 0 {
		receiverName = recv.Names[0].Name
	}

	// Get receiver type
	switch t := recv.Type.(type) {
	case *ast.Ident:
		receiverType = t.Name
	case *ast.StarExpr:
		receiverIsPtr = true
		if ident, ok := t.X.(*ast.Ident); ok {
			receiverType = ident.Name
		}
	}

	// Skip methods on unexported types
	if !isExported(receiverType) {
		return nil, "", nil
	}

	parsed := &ParsedMethod{
		Name:          fn.Name.Name,
		ReceiverName:  receiverName,
		ReceiverType:  receiverType,
		ReceiverIsPtr: receiverIsPtr,
	}

	if fn.Doc != nil {
		parsed.Doc = fn.Doc.Text()
	}

	// Parse parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			pt, err := p.parseType(field.Type)
			if err != nil {
				return nil, "", err
			}

			if _, isEllipsis := field.Type.(*ast.Ellipsis); isEllipsis {
				parsed.IsVariadic = true
			}

			if len(field.Names) == 0 {
				parsed.Params = append(parsed.Params, ParsedParam{Type: pt})
			} else {
				for _, name := range field.Names {
					parsed.Params = append(parsed.Params, ParsedParam{
						Name: name.Name,
						Type: pt,
					})
				}
			}
		}
	}

	// Parse results
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			pt, err := p.parseType(field.Type)
			if err != nil {
				return nil, "", err
			}

			if len(field.Names) == 0 {
				parsed.Results = append(parsed.Results, ParsedResult{Type: pt})
			} else {
				for _, name := range field.Names {
					parsed.Results = append(parsed.Results, ParsedResult{
						Name: name.Name,
						Type: pt,
					})
				}
			}
		}
	}

	return parsed, receiverType, nil
}

// parseStruct extracts struct information from ast.TypeSpec
func (p *Parser) parseStruct(ts *ast.TypeSpec, st *ast.StructType, doc *ast.CommentGroup) (*ParsedStruct, error) {
	// Skip unexported structs
	if !isExported(ts.Name.Name) {
		return nil, nil
	}

	parsed := &ParsedStruct{
		Name: ts.Name.Name,
	}

	if doc != nil {
		parsed.Doc = doc.Text()
	} else if ts.Doc != nil {
		parsed.Doc = ts.Doc.Text()
	}

	// Parse fields
	if st.Fields != nil {
		for _, field := range st.Fields.List {
			pt, err := p.parseType(field.Type)
			if err != nil {
				// Skip fields with unsupported types
				if p.verbose {
					fmt.Printf("Skipping field in %s: %v\n", ts.Name.Name, err)
				}
				continue
			}

			var tag string
			if field.Tag != nil {
				tag = field.Tag.Value
			}

			if len(field.Names) == 0 {
				// Embedded field
				parsed.Fields = append(parsed.Fields, ParsedField{
					Name:     pt.Name,
					Type:     pt,
					Tag:      tag,
					Exported: isExported(pt.Name),
				})
			} else {
				for _, name := range field.Names {
					parsed.Fields = append(parsed.Fields, ParsedField{
						Name:     name.Name,
						Type:     pt,
						Tag:      tag,
						Exported: isExported(name.Name),
					})
				}
			}
		}
	}

	return parsed, nil
}

// parseType converts ast.Expr to ParsedType
func (p *Parser) parseType(expr ast.Expr) (ParsedType, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.identToType(t.Name), nil

	case *ast.StarExpr:
		elem, err := p.parseType(t.X)
		if err != nil {
			return ParsedType{}, err
		}
		return ParsedType{
			Kind:      KindPointer,
			Name:      "*" + elem.Name,
			ElemType:  &elem,
			IsPointer: true,
		}, nil

	case *ast.ArrayType:
		elem, err := p.parseType(t.Elt)
		if err != nil {
			return ParsedType{}, err
		}
		if t.Len == nil {
			// Slice
			return ParsedType{
				Kind:     KindSlice,
				Name:     "[]" + elem.Name,
				ElemType: &elem,
			}, nil
		}
		// Fixed array
		size := getArrayLen(t.Len)
		return ParsedType{
			Kind:     KindArray,
			Name:     fmt.Sprintf("[%d]%s", size, elem.Name),
			ElemType: &elem,
			Size:     size,
		}, nil

	case *ast.MapType:
		key, err := p.parseType(t.Key)
		if err != nil {
			return ParsedType{}, err
		}
		val, err := p.parseType(t.Value)
		if err != nil {
			return ParsedType{}, err
		}
		return ParsedType{
			Kind:     KindMap,
			Name:     fmt.Sprintf("map[%s]%s", key.Name, val.Name),
			KeyType:  &key,
			ElemType: &val,
		}, nil

	case *ast.ChanType:
		return ParsedType{}, &UnsupportedTypeError{
			Type:   "chan",
			Reason: "channels cannot be exposed via CGO",
		}

	case *ast.InterfaceType:
		// Only support empty interface (interface{} / any)
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return ParsedType{
				Kind: KindInterface,
				Name: "interface{}",
			}, nil
		}
		return ParsedType{}, &UnsupportedTypeError{
			Type:   "interface",
			Reason: "non-empty interfaces cannot be exposed via CGO",
		}

	case *ast.SelectorExpr:
		// Imported type like pkg.Type
		if ident, ok := t.X.(*ast.Ident); ok {
			return ParsedType{
				Kind:        KindStruct, // Assume struct for imported types
				Name:        t.Sel.Name,
				PackagePath: ident.Name,
			}, nil
		}
		return ParsedType{}, &UnsupportedTypeError{
			Type:   "selector",
			Reason: "complex selector expressions not supported",
		}

	case *ast.FuncType:
		return ParsedType{}, &UnsupportedTypeError{
			Type:   "func",
			Reason: "function types cannot be exposed via CGO",
		}

	case *ast.Ellipsis:
		// Variadic parameter - parse the element type
		elem, err := p.parseType(t.Elt)
		if err != nil {
			return ParsedType{}, err
		}
		return ParsedType{
			Kind:     KindSlice,
			Name:     "..." + elem.Name,
			ElemType: &elem,
		}, nil

	default:
		return ParsedType{}, &UnsupportedTypeError{
			Type:   fmt.Sprintf("%T", expr),
			Reason: "unknown type expression",
		}
	}
}

// identToType converts a type name to ParsedType
func (p *Parser) identToType(name string) ParsedType {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"bool", "byte", "rune", "uintptr":
		return ParsedType{Kind: KindPrimitive, Name: name}
	case "string":
		return ParsedType{Kind: KindString, Name: name}
	case "error":
		return ParsedType{Kind: KindError, Name: name}
	default:
		// Assume it's a struct type defined in the same package
		return ParsedType{Kind: KindStruct, Name: name}
	}
}

// getArrayLen extracts array length from ast expression
func getArrayLen(expr ast.Expr) int {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		n, _ := strconv.Atoi(lit.Value)
		return n
	}
	return 0
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}
