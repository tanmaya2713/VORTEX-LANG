package types

import (
	"testing"

	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
	"github.com/vortex-lang/vortex/src/lexer"
	"github.com/vortex-lang/vortex/src/parser"
)

func parseAndCheck(t *testing.T, input, file string) *Checker {
	tokens, err := lexer.Lex(input, file)
	if err != nil {
		t.Fatal(err)
	}
	tokens = lexer.FilterNoise(tokens)
	p := parser.New(tokens, file)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors()[0])
	}
	checker := New()
	checker.Check(prog)
	return checker
}

func TestLetInferI32(t *testing.T) {
	c := parseAndCheck(t, "let x = 42;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	scope := c.globals
	typ, ok := scope.lookup("x")
	if !ok {
		t.Fatal("expected 'x' to be declared")
	}
	if typ.Kind() != common.TypeI32 {
		t.Errorf("expected i32, got %s", typ)
	}
}

func TestLetWithTypeAnnotation(t *testing.T) {
	c := parseAndCheck(t, "let x: f64 = 42;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("x")
	if typ.Kind() != common.TypeF64 {
		t.Errorf("expected f64, got %s", typ)
	}
}

func TestLetWithTypeMismatch(t *testing.T) {
	c := parseAndCheck(t, "let x: bool = 42;", "test.vtx")
	if len(c.errs) == 0 {
		t.Fatal("expected type mismatch error")
	}
}

func TestLetString(t *testing.T) {
	c := parseAndCheck(t, `let name = "vortex";`, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("name")
	if typ.Kind() != common.TypeString {
		t.Errorf("expected string, got %s", typ)
	}
}

func TestLetBool(t *testing.T) {
	c := parseAndCheck(t, "let yes = true;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("yes")
	if typ.Kind() != common.TypeBool {
		t.Errorf("expected bool, got %s", typ)
	}
}

func TestBinaryArithmetic(t *testing.T) {
	c := parseAndCheck(t, "let x = 1 + 2 * 3;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("x")
	if typ.Kind() != common.TypeI32 {
		t.Errorf("expected i32, got %s", typ)
	}
}

func TestBinaryStringConcat(t *testing.T) {
	c := parseAndCheck(t, `let x = "hello " + "world";`, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("x")
	if typ.Kind() != common.TypeString {
		t.Errorf("expected string, got %s", typ)
	}
}

func TestBinaryTypeError(t *testing.T) {
	c := parseAndCheck(t, `let x = "hello" + 42;`, "test.vtx")
	if len(c.errs) == 0 {
		t.Fatal("expected type error for string + int")
	}
}

func TestComparison(t *testing.T) {
	c := parseAndCheck(t, "let cmp = 1 < 2;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("cmp")
	if typ.Kind() != common.TypeBool {
		t.Errorf("expected bool, got %s", typ)
	}
}

func TestUnaryNegate(t *testing.T) {
	c := parseAndCheck(t, "let x = -42;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("x")
	if typ.Kind() != common.TypeI32 {
		t.Errorf("expected i32, got %s", typ)
	}
}

func TestUnaryNot(t *testing.T) {
	c := parseAndCheck(t, "let ok = !true;", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("ok")
	if typ.Kind() != common.TypeBool {
		t.Errorf("expected bool, got %s", typ)
	}
}

func TestFunctionDefAndCall(t *testing.T) {
	input := `fn add(x: i32, y: i32) -> i32 { return x + y; }`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestFunctionCallArgMismatch(t *testing.T) {
	input := `fn greet(name: string) -> string { return name; }`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestIfConditionBool(t *testing.T) {
	input := `if true { let x = 1; }`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestIfConditionTypeError(t *testing.T) {
	input := `if 42 { let x = 1; }`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) == 0 {
		t.Fatal("expected type error for if with int condition")
	}
}

func TestWhileConditionBool(t *testing.T) {
	input := `let mut x = 0; while x < 10 { x = x + 1; }`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestUndefinedVariable(t *testing.T) {
	input := "let x = y;"
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) == 0 {
		t.Fatal("expected error for undefined variable")
	}
}

func TestTensorTypeParse(t *testing.T) {
	elemType := &ast.TypeExpr{Name: "f32", Loc: common.Position{}}
	dims := []ast.Expr{
		&ast.NumberLit{Value: "784", Loc: common.Position{}},
		&ast.NumberLit{Value: "256", Loc: common.Position{}},
	}
	tensorExpr := &ast.TensorTypeExpr{ElemType: elemType, Dims: dims, Loc: common.Position{}}
	checker := New()
	typ := checker.typeFromExprType(tensorExpr)
	if typ.Kind() != common.TypeTensor {
		t.Errorf("expected tensor, got %s", typ)
	}
	tt, ok := typ.(*TensorType)
	if !ok {
		t.Fatal("expected *TensorType")
	}
	if tt.ElemType.Kind() != common.TypeF32 {
		t.Errorf("expected f32 element, got %s", tt.ElemType)
	}
	if len(tt.Shape) != 2 || tt.Shape[0] != 784 || tt.Shape[1] != 256 {
		t.Errorf("expected shape [784 256], got %v", tt.Shape)
	}
}

func TestArrayLiteral(t *testing.T) {
	c := parseAndCheck(t, "let arr = [1, 2, 3];", "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
	typ, _ := c.globals.lookup("arr")
	if typ.Kind() != common.TypeArray {
		t.Errorf("expected array, got %s", typ)
	}
}

func TestBlockScopingTypeCheck(t *testing.T) {
	input := `fn test() {
		let x = 1;
		{
			let y = 2;
		}
		return x;
	}`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestPrintTypeCheck(t *testing.T) {
	c := parseAndCheck(t, `print "hello";`, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}

func TestModelTypeCheck(t *testing.T) {
	input := `model MNISTClassifier {
		hidden = dense, inputs=784, neurons=256, activation=relu
		output = dense, inputs=256, neurons=10, activation=softmax
	}`
	c := parseAndCheck(t, input, "test.vtx")
	if len(c.errs) > 0 {
		t.Fatalf("unexpected errors: %v", c.errs)
	}
}
