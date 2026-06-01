package types

import (
	"fmt"

	"github.com/vortex-lang/vortex/src/common"
)

type Type interface {
	Kind() common.DataType
	String() string
}

type Primitive struct {
	kind common.DataType
}

func (p *Primitive) Kind() common.DataType { return p.kind }
func (p *Primitive) String() string        { return p.kind.String() }

var (
	TypeI8     = &Primitive{common.TypeI8}
	TypeI16    = &Primitive{common.TypeI16}
	TypeI32    = &Primitive{common.TypeI32}
	TypeI64    = &Primitive{common.TypeI64}
	TypeU8     = &Primitive{common.TypeU8}
	TypeU16    = &Primitive{common.TypeU16}
	TypeU32    = &Primitive{common.TypeU32}
	TypeU64    = &Primitive{common.TypeU64}
	TypeF32    = &Primitive{common.TypeF32}
	TypeF64    = &Primitive{common.TypeF64}
	TypeBool   = &Primitive{common.TypeBool}
	TypeString = &Primitive{common.TypeString}
	TypeVoid   = &Primitive{common.TypeVoid}
	TypeError  = &Primitive{common.TypeError}
)

type FnType struct {
	ParamTypes []Type
	ReturnType Type
}

func (f *FnType) Kind() common.DataType { return common.TypeFn }
func (f *FnType) String() string {
	s := "fn("
	for i, p := range f.ParamTypes {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	s += ") -> " + f.ReturnType.String()
	return s
}

type StructType struct {
	Name   string
	Fields map[string]Type
}

func (s *StructType) Kind() common.DataType { return common.TypeStruct }
func (s *StructType) String() string        { return "struct " + s.Name }

type ArrayType struct {
	ElemType Type
	Len      int64
}

func (a *ArrayType) Kind() common.DataType { return common.TypeArray }
func (a *ArrayType) String() string {
	if a.Len > 0 {
		return fmt.Sprintf("[%d]%s", a.Len, a.ElemType)
	}
	return "[]" + a.ElemType.String()
}

type TensorType struct {
	ElemType Type
	Shape    []int64
}

func (t *TensorType) Kind() common.DataType { return common.TypeTensor }
func (t *TensorType) String() string {
	s := "tensor<" + t.ElemType.String()
	if len(t.Shape) > 0 {
		s += ", ["
		for i, d := range t.Shape {
			if i > 0 {
				s += ", "
			}
			if d < 0 {
				s += "?"
			} else {
				s += fmt.Sprintf("%d", d)
			}
		}
		s += "]"
	}
	s += ">"
	return s
}

type ModelType struct {
	Name string
}

func (m *ModelType) Kind() common.DataType { return common.TypeModel }
func (m *ModelType) String() string        { return "model " + m.Name }

type LayerType struct {
	LayerKind string
}

func (l *LayerType) Kind() common.DataType { return common.TypeLayer }
func (l *LayerType) String() string        { return "layer " + l.LayerKind }

type PtrType struct {
	ElemType Type
}

func (p *PtrType) Kind() common.DataType { return common.TypePtr }
func (p *PtrType) String() string        { return "*" + p.ElemType.String() }

type RefType struct {
	ElemType Type
}

func (r *RefType) Kind() common.DataType { return common.TypeRef }
func (r *RefType) String() string        { return "ref " + r.ElemType.String() }

type NamedType struct {
	Name string
}

func (n *NamedType) Kind() common.DataType { return common.TypeStruct }
func (n *NamedType) String() string        { return n.Name }
