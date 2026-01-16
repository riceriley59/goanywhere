package core

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Suite")
}

var _ = Describe("UnsupportedTypeError", func() {
	It("formats error message correctly", func() {
		err := &UnsupportedTypeError{
			Type:   "chan",
			Reason: "channels cannot be exposed",
		}
		Expect(err.Error()).To(Equal("unsupported type chan: channels cannot be exposed"))
	})
})

var _ = Describe("TypeKind constants", func() {
	It("has correct values", func() {
		Expect(KindPrimitive).To(Equal(TypeKind(0)))
		Expect(KindString).To(Equal(TypeKind(1)))
		Expect(KindStruct).To(Equal(TypeKind(2)))
		Expect(KindSlice).To(Equal(TypeKind(3)))
		Expect(KindArray).To(Equal(TypeKind(4)))
		Expect(KindMap).To(Equal(TypeKind(5)))
		Expect(KindPointer).To(Equal(TypeKind(6)))
		Expect(KindInterface).To(Equal(TypeKind(7)))
		Expect(KindFunc).To(Equal(TypeKind(8)))
		Expect(KindChan).To(Equal(TypeKind(9)))
		Expect(KindError).To(Equal(TypeKind(10)))
	})
})

var _ = Describe("Parser", func() {
	var parser *Parser

	BeforeEach(func() {
		parser = NewParser(false)
	})

	Describe("NewParser", func() {
		It("creates parser with verbose off", func() {
			p := NewParser(false)
			Expect(p).NotTo(BeNil())
			Expect(p.verbose).To(BeFalse())
		})

		It("creates parser with verbose on", func() {
			p := NewParser(true)
			Expect(p.verbose).To(BeTrue())
		})
	})

	Describe("ParsePackage", func() {
		var fixtureDir string

		BeforeEach(func() {
			wd, _ := os.Getwd()
			fixtureDir = filepath.Join(wd, "..", "..", "tests", "fixtures", "simple")
		})

		It("parses exported functions", func() {
			pkg, err := parser.ParsePackage(fixtureDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg.Name).To(Equal("simple"))
			Expect(len(pkg.Functions)).To(BeNumerically(">=", 5))

			var addFn *ParsedFunc
			for i := range pkg.Functions {
				if pkg.Functions[i].Name == "Add" {
					addFn = &pkg.Functions[i]
					break
				}
			}
			Expect(addFn).NotTo(BeNil())
			Expect(addFn.Params).To(HaveLen(2))
			Expect(addFn.Results).To(HaveLen(1))
		})

		It("parses exported structs", func() {
			pkg, err := parser.ParsePackage(fixtureDir)
			Expect(err).NotTo(HaveOccurred())

			var point *ParsedStruct
			for i := range pkg.Structs {
				if pkg.Structs[i].Name == "Point" {
					point = &pkg.Structs[i]
					break
				}
			}
			Expect(point).NotTo(BeNil())
			Expect(point.Fields).To(HaveLen(2))
		})

		It("parses methods on structs", func() {
			pkg, err := parser.ParsePackage(fixtureDir)
			Expect(err).NotTo(HaveOccurred())

			var point *ParsedStruct
			for i := range pkg.Structs {
				if pkg.Structs[i].Name == "Point" {
					point = &pkg.Structs[i]
					break
				}
			}
			Expect(point).NotTo(BeNil())
			Expect(len(point.Methods)).To(BeNumerically(">=", 2))
		})

		It("skips unexported functions", func() {
			pkg, err := parser.ParsePackage(fixtureDir)
			Expect(err).NotTo(HaveOccurred())

			for _, fn := range pkg.Functions {
				Expect(fn.Name).NotTo(Equal("unexported"))
			}
		})

		It("returns error for non-existent directory", func() {
			_, err := parser.ParsePackage("/nonexistent/path")
			Expect(err).To(HaveOccurred())
		})

		It("returns error for empty directory", func() {
			tmpDir, _ := os.MkdirTemp("", "empty")
			defer func() { _ = os.RemoveAll(tmpDir) }()

			_, err := parser.ParsePackage(tmpDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("identToType", func() {
		It("maps primitive types", func() {
			primitives := []string{"int", "int8", "int16", "int32", "int64",
				"uint", "uint8", "uint16", "uint32", "uint64",
				"float32", "float64", "bool", "byte", "rune", "uintptr"}

			for _, p := range primitives {
				pt := parser.identToType(p)
				Expect(pt.Kind).To(Equal(KindPrimitive))
				Expect(pt.Name).To(Equal(p))
			}
		})

		It("maps string type", func() {
			pt := parser.identToType("string")
			Expect(pt.Kind).To(Equal(KindString))
		})

		It("maps error type", func() {
			pt := parser.identToType("error")
			Expect(pt.Kind).To(Equal(KindError))
		})

		It("maps unknown types as struct", func() {
			pt := parser.identToType("MyCustomType")
			Expect(pt.Kind).To(Equal(KindStruct))
		})
	})

	Describe("isExported", func() {
		It("returns true for uppercase names", func() {
			Expect(isExported("Foo")).To(BeTrue())
			Expect(isExported("FOO")).To(BeTrue())
		})

		It("returns false for lowercase names", func() {
			Expect(isExported("foo")).To(BeFalse())
			Expect(isExported("_foo")).To(BeFalse())
		})

		It("returns false for empty string", func() {
			Expect(isExported("")).To(BeFalse())
		})
	})
})
