package ast

import "fmt"

type Printer struct {
	indent int
}

func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) Indent() {
	for i := 0; i < p.indent; i++ {
		fmt.Print("  ")
	}
}

func (p *Printer) Print(prog *Program) {
	fmt.Println("(program")
	p.indent++
	for _, stmt := range prog.Stmts {
		p.printStmt(stmt)
	}
	p.indent--
	fmt.Println(")")
}

func (p *Printer) printStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *LetStmt:
		p.Indent()
		mut := ""
		if s.Mut {
			mut = " mut"
		}
		fmt.Printf("(let%s %s", mut, s.Name)
		if s.Type != nil {
			fmt.Printf(" : %s", s.Type)
		}
		fmt.Printf(" = ")
		p.printExpr(s.Value)
		fmt.Println(")")
	case *FnDef:
		p.Indent()
		fmt.Printf("(fn %s\n", s.Name)
		p.indent++
		for _, param := range s.Params {
			p.Indent()
			fmt.Printf("(param %s", param.Name)
			if param.Type != nil {
				fmt.Printf(" : %s", param.Type)
			}
			fmt.Println(")")
		}
		p.printBlock(s.Body)
		p.indent--
		p.Indent()
		fmt.Println(")")
	case *IfStmt:
		p.Indent()
		fmt.Print("(if ")
		p.printExpr(s.Cond)
		fmt.Println()
		p.printBlock(s.Then)
		if s.Else != nil {
			p.Indent()
			fmt.Println("(else")
			p.printStmt(s.Else)
			p.Indent()
			fmt.Println(")")
		}
		p.Indent()
		fmt.Println(")")
	case *ForStmt:
		p.Indent()
		fmt.Printf("(for %s in ", s.Var)
		p.printExpr(s.Range)
		fmt.Println()
		p.printBlock(s.Body)
		p.Indent()
		fmt.Println(")")
	case *WhileStmt:
		p.Indent()
		fmt.Print("(while ")
		p.printExpr(s.Cond)
		fmt.Println()
		p.printBlock(s.Body)
		p.Indent()
		fmt.Println(")")
	case *ReturnStmt:
		p.Indent()
		if s.Value != nil {
			fmt.Print("(return ")
			p.printExpr(s.Value)
			fmt.Println(")")
		} else {
			fmt.Println("(return)")
		}
	case *BreakStmt:
		p.Indent()
		fmt.Println("(break)")
	case *ContinueStmt:
		p.Indent()
		fmt.Println("(continue)")
	case *ImportStmt:
		p.Indent()
		fmt.Printf("(import %s)\n", s.Path)
	case *PrintStmt:
		p.Indent()
		fmt.Print("(print ")
		p.printExpr(s.Expr)
		fmt.Println(")")
	case *AssertStmt:
		p.Indent()
		fmt.Print("(assert ")
		p.printExpr(s.Cond)
		if s.Msg != nil {
			fmt.Print(" ")
			p.printExpr(s.Msg)
		}
		fmt.Println(")")
	case *ExprStmt:
		p.Indent()
		p.printExpr(s.E)
		fmt.Println()
	case *StructDef:
		p.Indent()
		fmt.Printf("(struct %s\n", s.Name)
		p.indent++
		for _, f := range s.Fields {
			p.Indent()
			fmt.Printf("(field %s : %s)\n", f.Name, f.Type)
		}
		p.indent--
		p.Indent()
		fmt.Println(")")
	case *ModelDef:
		p.Indent()
		fmt.Printf("(model %s\n", s.Name)
		p.indent++
		for _, l := range s.Layers {
			p.Indent()
			fmt.Printf("(layer %s = %s", l.Name, l.Kind)
			for _, np := range l.Params {
				fmt.Printf(" %s=%s", np.Name, np.Value)
			}
			fmt.Println(")")
		}
		p.indent--
		p.Indent()
		fmt.Println(")")
	case *TrainStmt:
		p.Indent()
		fmt.Printf("(train %s", s.Model)
		if s.Data != nil {
			fmt.Printf(" data=%s", s.Data)
		}
		if s.Epochs != nil {
			fmt.Printf(" epochs=%s", s.Epochs)
		}
		if s.LR != nil {
			fmt.Printf(" lr=%s", s.LR)
		}
		fmt.Println(")")
	}
}

func (p *Printer) printBlock(block *BlockStmt) {
	p.indent++
	for _, stmt := range block.Stmts {
		p.printStmt(stmt)
	}
	p.indent--
}

func (p *Printer) printExpr(expr Expr) {
	switch e := expr.(type) {
	case *Ident:
		fmt.Print(e.Name)
	case *NumberLit:
		fmt.Print(e.Value)
	case *StringLit:
		fmt.Printf("%q", e.Value)
	case *BoolLit:
		fmt.Print(e.Value)
	case *BinaryExpr:
		fmt.Print("(")
		p.printExpr(e.Left)
		fmt.Printf(" %s ", e.Op)
		p.printExpr(e.Right)
		fmt.Print(")")
	case *UnaryExpr:
		fmt.Printf("(%s", e.Op)
		p.printExpr(e.Right)
		fmt.Print(")")
	case *CallExpr:
		p.printExpr(e.Fn)
		fmt.Print("(")
		for i, arg := range e.Args {
			if i > 0 {
				fmt.Print(" ")
			}
			p.printExpr(arg)
		}
		fmt.Print(")")
	case *AssignExpr:
		fmt.Print("(= ")
		p.printExpr(e.Left)
		fmt.Print(" ")
		p.printExpr(e.Right)
		fmt.Print(")")
	case *TypeExpr:
		fmt.Print(e.Name)
	case *IndexExpr:
		p.printExpr(e.Target)
		fmt.Print("[")
		p.printExpr(e.Index)
		fmt.Print("]")
	case *MemberExpr:
		p.printExpr(e.Target)
		fmt.Print(".")
		fmt.Print(e.Name)
	case *TensorTypeExpr:
		fmt.Print("tensor<")
		p.printExpr(e.ElemType)
		if len(e.Dims) > 0 {
			fmt.Print(", ")
			for i, d := range e.Dims {
				if i > 0 {
					fmt.Print(", ")
				}
				if d == nil {
					fmt.Print("?")
				} else {
					p.printExpr(d)
				}
			}
		}
		fmt.Print(">")
	case *ArrayLit:
		fmt.Print("[")
		for i, el := range e.Elems {
			if i > 0 {
				fmt.Print(", ")
			}
			p.printExpr(el)
		}
		fmt.Print("]")
	default:
		fmt.Print("?")
	}
}
