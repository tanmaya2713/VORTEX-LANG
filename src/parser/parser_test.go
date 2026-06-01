package parser

import (
	"testing"

	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/lexer"
)

func parseInput(t *testing.T, input, file string) *ast.Program {
	tokens, err := lexer.Lex(input, file)
	if err != nil {
		t.Fatal(err)
	}
	tokens = lexer.FilterNoise(tokens)
	p := New(tokens, file)
	prog := p.Parse()
	if len(p.errs) > 0 {
		t.Fatalf("parse errors: %v", p.errs[0])
	}
	return prog
}

func TestLetStatement(t *testing.T) {
	prog := parseInput(t, "let x = 42;", "test.vtx")
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Stmts))
	}
	letStmt, ok := prog.Stmts[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("expected LetStmt, got %T", prog.Stmts[0])
	}
	if letStmt.Name.String() != "x" {
		t.Errorf("expected name 'x', got '%s'", letStmt.Name)
	}
	if letStmt.Mut {
		t.Errorf("expected non-mutable let")
	}
}

func TestLetMut(t *testing.T) {
	prog := parseInput(t, "let mut count = 0;", "test.vtx")
	letStmt := prog.Stmts[0].(*ast.LetStmt)
	if !letStmt.Mut {
		t.Errorf("expected mutable let")
	}
}

func TestFunctionDef(t *testing.T) {
	input := `fn add(x: i32, y: i32) -> i32 { return x + y; }`
	prog := parseInput(t, input, "test.vtx")
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Stmts))
	}
	fn, ok := prog.Stmts[0].(*ast.FnDef)
	if !ok {
		t.Fatalf("expected FnDef, got %T", prog.Stmts[0])
	}
	if fn.Name.String() != "add" {
		t.Errorf("expected function name 'add', got '%s'", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
}

func TestIfElse(t *testing.T) {
	input := `if x > 0 { return x; } else { return 0; }`
	prog := parseInput(t, input, "test.vtx")
	ifStmt, ok := prog.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", prog.Stmts[0])
	}
	if ifStmt.Else == nil {
		t.Errorf("expected else branch")
	}
}

func TestForLoop(t *testing.T) {
	input := `let sum = 0;
for i in range(10) {
    sum = sum + i;
}`
	prog := parseInput(t, input, "test.vtx")
	if len(prog.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Stmts))
	}
	_, ok := prog.Stmts[1].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", prog.Stmts[1])
	}
}

func TestWhileLoop(t *testing.T) {
	input := `let mut x = 0;
while x < 10 {
    x = x + 1;
}`
	prog := parseInput(t, input, "test.vtx")
	if len(prog.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Stmts))
	}
	_, ok := prog.Stmts[1].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", prog.Stmts[1])
	}
}

func TestReturnStatement(t *testing.T) {
	input := `fn answer() -> i32 { return 42; }`
	prog := parseInput(t, input, "test.vtx")
	fn := prog.Stmts[0].(*ast.FnDef)
	retStmt, ok := fn.Body.Stmts[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", fn.Body.Stmts[0])
	}
	if retStmt.Value == nil {
		t.Errorf("expected return value")
	}
}

func TestModelDefinition(t *testing.T) {
	input := `model MNISTClassifier {
    hidden = dense, inputs=784, neurons=256, activation=relu
    output = dense, inputs=256, neurons=10, activation=softmax
}`
	prog := parseInput(t, input, "test.vtx")
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Stmts))
	}
	model, ok := prog.Stmts[0].(*ast.ModelDef)
	if !ok {
		t.Fatalf("expected ModelDef, got %T", prog.Stmts[0])
	}
	if model.Name.String() != "MNISTClassifier" {
		t.Errorf("expected model name 'MNISTClassifier', got '%s'", model.Name)
	}
	if len(model.Layers) != 2 {
		t.Errorf("expected 2 layers, got %d", len(model.Layers))
	}
}

func TestOperatorPrecedence(t *testing.T) {
	input := `let x = 1 + 2 * 3;`
	prog := parseInput(t, input, "test.vtx")
	letStmt := prog.Stmts[0].(*ast.LetStmt)
	bin, ok := letStmt.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", letStmt.Value)
	}
	if bin.Op != "+" {
		t.Errorf("expected '+', got '%s'", bin.Op)
	}
	right, ok := bin.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right BinaryExpr for *, got %T", bin.Right)
	}
	if right.Op != "*" {
		t.Errorf("expected '*', got '%s'", right.Op)
	}
}

func TestFunctionCall(t *testing.T) {
	input := `add(3, 4);`
	prog := parseInput(t, input, "test.vtx")
	exprStmt, ok := prog.Stmts[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", prog.Stmts[0])
	}
	call, ok := exprStmt.E.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.E)
	}
	if len(call.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(call.Args))
	}
}

func TestBlockScoping(t *testing.T) {
	input := `fn test() {
    let x = 1;
    {
        let y = 2;
    }
    return x;
}`
	prog := parseInput(t, input, "test.vtx")
	fn := prog.Stmts[0].(*ast.FnDef)
	if len(fn.Body.Stmts) != 3 {
		t.Errorf("expected 3 statements in body, got %d", len(fn.Body.Stmts))
	}
}

func TestPrint(t *testing.T) {
	input := `print "hello";`
	prog := parseInput(t, input, "test.vtx")
	_, ok := prog.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected PrintStmt, got %T", prog.Stmts[0])
	}
}

func TestNoiseFiltered(t *testing.T) {
	input := `let x the is to be = 42 and this or that;`
	tokens, err := lexer.Lex(input, "test.vtx")
	if err != nil {
		t.Fatal(err)
	}
	tokens = lexer.FilterNoise(tokens)
	p := New(tokens, "test.vtx")
	prog := p.Parse()
	if len(p.errs) > 0 {
		t.Fatalf("parse errors after noise filtering: %v", p.errs[0])
	}
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Stmts))
	}
}

func TestEmptyProgram(t *testing.T) {
	prog := parseInput(t, "", "test.vtx")
	if len(prog.Stmts) != 0 {
		t.Errorf("expected 0 statements for empty program, got %d", len(prog.Stmts))
	}
}

func TestLetWithType(t *testing.T) {
	input := `let name: string = "vortex";`
	prog := parseInput(t, input, "test.vtx")
	letStmt := prog.Stmts[0].(*ast.LetStmt)
	if letStmt.Type == nil {
		t.Errorf("expected type annotation")
	}
}
