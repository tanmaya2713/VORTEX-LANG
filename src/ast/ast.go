package ast

import (
	"github.com/vortex-lang/vortex/src/common"
)

type Node interface {
	Pos() common.Position
	String() string
}

type Program struct {
	Stmts []Stmt
}

func (p *Program) Pos() common.Position {
	if len(p.Stmts) > 0 {
		return p.Stmts[0].Pos()
	}
	return common.Position{}
}

func (p *Program) String() string { return "Program" }

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type Ident struct {
	Name string
	Loc  common.Position
}

func (i *Ident) Pos() common.Position { return i.Loc }
func (i *Ident) String() string       { return i.Name }
func (i *Ident) exprNode()            {}
func (i *Ident) stmtNode()            {}

type NumberLit struct {
	Value string
	Kind  string
	Loc   common.Position
}

func (n *NumberLit) Pos() common.Position { return n.Loc }
func (n *NumberLit) String() string       { return n.Value }
func (n *NumberLit) exprNode()            {}

type StringLit struct {
	Value string
	Loc   common.Position
}

func (s *StringLit) Pos() common.Position { return s.Loc }
func (s *StringLit) String() string       { return "\"" + s.Value + "\"" }
func (s *StringLit) exprNode()            {}

type BoolLit struct {
	Value bool
	Loc   common.Position
}

func (b *BoolLit) Pos() common.Position { return b.Loc }
func (b *BoolLit) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (b *BoolLit) exprNode() {}

type LetStmt struct {
	Name  *Ident
	Type  Expr
	Value Expr
	Mut   bool
	Loc   common.Position
}

func (l *LetStmt) Pos() common.Position { return l.Loc }
func (l *LetStmt) String() string       { return "Let " + l.Name.String() }
func (l *LetStmt) stmtNode()            {}

type FnDef struct {
	Name   *Ident
	Params []*ParamDef
	Body   *BlockStmt
	Return Expr
	Loc    common.Position
}

func (f *FnDef) Pos() common.Position { return f.Loc }
func (f *FnDef) String() string       { return "Fn " + f.Name.String() }
func (f *FnDef) stmtNode()            {}

type ParamDef struct {
	Name *Ident
	Type Expr
	Loc  common.Position
}

func (p *ParamDef) Pos() common.Position { return p.Loc }
func (p *ParamDef) String() string       { return p.Name.String() }

type BlockStmt struct {
	Stmts []Stmt
	Loc   common.Position
}

func (b *BlockStmt) Pos() common.Position { return b.Loc }
func (b *BlockStmt) String() string       { return "Block" }
func (b *BlockStmt) stmtNode()            {}

type IfStmt struct {
	Cond Expr
	Then *BlockStmt
	Else Stmt
	Loc  common.Position
}

func (i *IfStmt) Pos() common.Position { return i.Loc }
func (i *IfStmt) String() string       { return "If" }
func (i *IfStmt) stmtNode()            {}

type ForStmt struct {
	Var   *Ident
	Range Expr
	Body  *BlockStmt
	Loc   common.Position
}

func (f *ForStmt) Pos() common.Position { return f.Loc }
func (f *ForStmt) String() string       { return "For" }
func (f *ForStmt) stmtNode()            {}

type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
	Loc  common.Position
}

func (w *WhileStmt) Pos() common.Position { return w.Loc }
func (w *WhileStmt) String() string       { return "While" }
func (w *WhileStmt) stmtNode()            {}

type ReturnStmt struct {
	Value Expr
	Loc   common.Position
}

func (r *ReturnStmt) Pos() common.Position { return r.Loc }
func (r *ReturnStmt) String() string       { return "Return" }
func (r *ReturnStmt) stmtNode()            {}

type BreakStmt struct {
	Loc common.Position
}

func (b *BreakStmt) Pos() common.Position { return b.Loc }
func (b *BreakStmt) String() string       { return "Break" }
func (b *BreakStmt) stmtNode()            {}

type ContinueStmt struct {
	Loc common.Position
}

func (c *ContinueStmt) Pos() common.Position { return c.Loc }
func (c *ContinueStmt) String() string       { return "Continue" }
func (c *ContinueStmt) stmtNode()            {}

type ExprStmt struct {
	E   Expr
	Loc common.Position
}

func (e *ExprStmt) Pos() common.Position { return e.Loc }
func (e *ExprStmt) String() string       { return e.E.String() }
func (e *ExprStmt) stmtNode()            {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
	Loc   common.Position
}

func (b *BinaryExpr) Pos() common.Position { return b.Loc }
func (b *BinaryExpr) String() string {
	return "(" + b.Left.String() + " " + b.Op + " " + b.Right.String() + ")"
}
func (b *BinaryExpr) exprNode() {}

type UnaryExpr struct {
	Op    string
	Right Expr
	Loc   common.Position
}

func (u *UnaryExpr) Pos() common.Position { return u.Loc }
func (u *UnaryExpr) String() string       { return "(" + u.Op + u.Right.String() + ")" }
func (u *UnaryExpr) exprNode()            {}

type CallExpr struct {
	Fn   Expr
	Args []Expr
	Loc  common.Position
}

func (c *CallExpr) Pos() common.Position { return c.Loc }
func (c *CallExpr) String() string       { return c.Fn.String() + "()" }
func (c *CallExpr) exprNode()            {}

type AssignExpr struct {
	Left  Expr
	Right Expr
	Loc   common.Position
}

func (a *AssignExpr) Pos() common.Position { return a.Loc }
func (a *AssignExpr) String() string       { return a.Left.String() + " = " + a.Right.String() }
func (a *AssignExpr) exprNode()            {}
func (a *AssignExpr) stmtNode()            {}

type TypeExpr struct {
	Name string
	Args []Expr
	Loc  common.Position
}

func (t *TypeExpr) Pos() common.Position { return t.Loc }
func (t *TypeExpr) String() string       { return t.Name }
func (t *TypeExpr) exprNode()            {}

type IndexExpr struct {
	Target Expr
	Index  Expr
	Loc    common.Position
}

func (i *IndexExpr) Pos() common.Position { return i.Loc }
func (i *IndexExpr) String() string       { return i.Target.String() + "[" + i.Index.String() + "]" }
func (i *IndexExpr) exprNode()            {}

type MemberExpr struct {
	Target Expr
	Name   *Ident
	Loc    common.Position
}

func (m *MemberExpr) Pos() common.Position { return m.Loc }
func (m *MemberExpr) String() string       { return m.Target.String() + "." + m.Name.String() }
func (m *MemberExpr) exprNode()            {}

type ImportStmt struct {
	Path *StringLit
	Loc  common.Position
}

func (i *ImportStmt) Pos() common.Position { return i.Loc }
func (i *ImportStmt) String() string       { return "Import " + i.Path.String() }
func (i *ImportStmt) stmtNode()            {}

type PrintStmt struct {
	Expr Expr
	Loc  common.Position
}

func (p *PrintStmt) Pos() common.Position { return p.Loc }
func (p *PrintStmt) String() string       { return "Print " + p.Expr.String() }
func (p *PrintStmt) stmtNode()            {}

type AssertStmt struct {
	Cond Expr
	Msg  Expr
	Loc  common.Position
}

func (a *AssertStmt) Pos() common.Position { return a.Loc }
func (a *AssertStmt) String() string       { return "Assert " + a.Cond.String() }
func (a *AssertStmt) stmtNode()            {}

type StructDef struct {
	Name   *Ident
	Fields []*FieldDef
	Loc    common.Position
}

func (s *StructDef) Pos() common.Position { return s.Loc }
func (s *StructDef) String() string       { return "Struct " + s.Name.String() }
func (s *StructDef) stmtNode()            {}

type FieldDef struct {
	Name *Ident
	Type Expr
	Loc  common.Position
}

func (f *FieldDef) Pos() common.Position { return f.Loc }
func (f *FieldDef) String() string       { return f.Name.String() }

type ModelDef struct {
	Name   *Ident
	Layers []*LayerDef
	Loc    common.Position
}

func (m *ModelDef) Pos() common.Position { return m.Loc }
func (m *ModelDef) String() string       { return "Model " + m.Name.String() }
func (m *ModelDef) stmtNode()            {}

type LayerDef struct {
	Name   *Ident
	Kind   Expr
	Params []*NamedParam
	Loc    common.Position
}

func (l *LayerDef) Pos() common.Position { return l.Loc }
func (l *LayerDef) String() string       { return l.Name.String() }

type NamedParam struct {
	Name  *Ident
	Value Expr
	Loc   common.Position
}

func (n *NamedParam) Pos() common.Position { return n.Loc }
func (n *NamedParam) String() string       { return n.Name.String() + "=" + n.Value.String() }

type TrainStmt struct {
	Model    Expr
	Data     Expr
	Epochs   Expr
	LR       Expr
	Strategy Expr
	Devices  Expr
	Loc      common.Position
}

func (t *TrainStmt) Pos() common.Position { return t.Loc }
func (t *TrainStmt) String() string       { return "Train " + t.Model.String() }
func (t *TrainStmt) stmtNode()            {}

type TensorTypeExpr struct {
	ElemType Expr
	Dims     []Expr
	Loc      common.Position
}

func (t *TensorTypeExpr) Pos() common.Position { return t.Loc }
func (t *TensorTypeExpr) String() string {
	s := "tensor<" + t.ElemType.String()
	for _, d := range t.Dims {
		if d != nil {
			s += ", " + d.String()
		}
	}
	s += ">"
	return s
}
func (t *TensorTypeExpr) exprNode() {}

type ArrayLit struct {
	Elems []Expr
	Loc   common.Position
}

func (a *ArrayLit) Pos() common.Position { return a.Loc }
func (a *ArrayLit) String() string       { return "Array" }
func (a *ArrayLit) exprNode()            {}
