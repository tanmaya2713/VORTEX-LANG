package llvmir

import (
	"strings"
	"testing"

	"github.com/vortex-lang/vortex/src/lexer"
	"github.com/vortex-lang/vortex/src/parser"
	"github.com/vortex-lang/vortex/src/types"
)

func testCodegen(source string) (string, []error) {
	tokens, err := lexer.Lex(source, "test")
	if err != nil {
		return "", []error{err}
	}
	prog, parseErrs := parser.ParseTokens(tokens, "test")
	if len(parseErrs) > 0 {
		errs := make([]error, len(parseErrs))
		for i, e := range parseErrs {
			errs[i] = e
		}
		return "", errs
	}
	checker := types.New()
	if !checker.Check(prog) {
		return "", checker.Errors()
	}
	c := New()
	mod := c.Compile(prog)
	if len(c.Errors()) > 0 {
		return mod.String(), c.Errors()
	}
	return mod.String(), nil
}

func parseAndCodegen(t *testing.T, source string) string {
	t.Helper()
	ir, errs := testCodegen(source)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		t.Fatalf("codegen errors: %s\nIR:\n%s", strings.Join(msgs, "; "), ir)
	}
	return ir
}

func TestCodegenLetI32(t *testing.T) {
	ir := parseAndCodegen(t, "let x: i32 = 42;")
	if !strings.Contains(ir, "store i32 42") {
		t.Errorf("expected store i32 42 in IR:\n%s", ir)
	}
}

func TestCodegenLetF64(t *testing.T) {
	ir := parseAndCodegen(t, "let x: f64 = 3.14;")
	if !strings.Contains(ir, "store double") {
		t.Errorf("expected store double in IR:\n%s", ir)
	}
}

func TestCodegenLetBool(t *testing.T) {
	ir := parseAndCodegen(t, "let x: bool = true;")
	if !strings.Contains(ir, "store i1 true") {
		t.Errorf("expected store i1 true in IR:\n%s", ir)
	}
}

func TestCodegenBinaryAdd(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 + 2;")
	if !strings.Contains(ir, "add i32") {
		t.Errorf("expected add i32 in IR:\n%s", ir)
	}
}

func TestCodegenBinarySub(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 - 2;")
	if !strings.Contains(ir, "sub i32") {
		t.Errorf("expected sub i32 in IR:\n%s", ir)
	}
}

func TestCodegenBinaryMul(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 * 2;")
	if !strings.Contains(ir, "mul i32") {
		t.Errorf("expected mul i32 in IR:\n%s", ir)
	}
}

func TestCodegenBinaryDiv(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 / 2;")
	if !strings.Contains(ir, "sdiv i32") {
		t.Errorf("expected sdiv i32 in IR:\n%s", ir)
	}
}

func TestCodegenBinaryFAdd(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1.0 + 2.0;")
	if !strings.Contains(ir, "fadd double") {
		t.Errorf("expected fadd double in IR:\n%s", ir)
	}
}

func TestCodegenBinaryEq(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 == 2;")
	if !strings.Contains(ir, "icmp eq i32") {
		t.Errorf("expected icmp eq i32 in IR:\n%s", ir)
	}
}

func TestCodegenBinaryLt(t *testing.T) {
	ir := parseAndCodegen(t, "let x = 1 < 2;")
	if !strings.Contains(ir, "icmp slt i32") {
		t.Errorf("expected icmp slt i32 in IR:\n%s", ir)
	}
}

func TestCodegenUnaryNeg(t *testing.T) {
	ir := parseAndCodegen(t, "let x = -5;")
	if !strings.Contains(ir, "sub i32 0,") {
		t.Errorf("expected sub for neg in IR:\n%s", ir)
	}
}

func TestCodegenUnaryBoolNot(t *testing.T) {
	ir := parseAndCodegen(t, "let x = !true;")
	if !strings.Contains(ir, "xor i1 true, true") {
		t.Errorf("expected xor for not in IR:\n%s", ir)
	}
}

func TestCodegenIf(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { if true { print(1); } }")
	if !strings.Contains(ir, "br i1") {
		t.Errorf("expected br i1 in IR:\n%s", ir)
	}
}

func TestCodegenWhile(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { while true { print(1); } }")
	if !strings.Contains(ir, "br label") {
		t.Errorf("expected br label in IR:\n%s", ir)
	}
}

func TestCodegenFnDef(t *testing.T) {
	ir := parseAndCodegen(t, "fn add(x: i32, y: i32) -> i32 { return x + y; }")
	if !strings.Contains(ir, "define i32 @add(i32") {
		t.Errorf("expected function definition in IR:\n%s", ir)
	}
}

func TestCodegenFnReturnVoid(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { return; }")
	if !strings.Contains(ir, "ret void") {
		t.Errorf("expected ret void in IR:\n%s", ir)
	}
}

func TestCodegenPrint(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { print(42); }")
	if !strings.Contains(ir, "call void @vortex_print_i32") {
		t.Errorf("expected call to vortex_print_i32 in IR:\n%s", ir)
	}
}

func TestCodegenPrintStr(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { print(\"hello\"); }")
	if !strings.Contains(ir, "call void @vortex_print_string") {
		t.Errorf("expected call to vortex_print_string in IR:\n%s", ir)
	}
}

func TestCodegenAssign(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { let mut x: i32 = 10; x = 20; }")
	if !strings.Contains(ir, "store i32 20") {
		t.Errorf("expected store i32 20 in IR:\n%s", ir)
	}
}

func TestCodegenStrConstant(t *testing.T) {
	ir := parseAndCodegen(t, "let s: string = \"hi\";")
	if !strings.Contains(ir, "c\"hi\\00\"") {
		t.Errorf("expected string constant c\"hi\\00\" in IR:\n%s", ir)
	}
}

func TestCodegenReturnI32(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() -> i32 { return 42; }")
	if !strings.Contains(ir, "ret i32 42") {
		t.Errorf("expected ret i32 42 in IR:\n%s", ir)
	}
}

func TestCodegenReturnVoid(t *testing.T) {
	ir := parseAndCodegen(t, "fn main() { return; }")
	if !strings.Contains(ir, "ret void") {
		t.Errorf("expected ret void in IR:\n%s", ir)
	}
}

func TestCodegenNestedBlock(t *testing.T) {
	ir := parseAndCodegen(t, `
		fn main() {
			let x: i32 = 1;
			{
				let y: i32 = 2;
				print(x + y);
			}
		}
	`)
	if !strings.Contains(ir, "add i32") {
		t.Errorf("expected add i32 in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "vortex_print_i32") {
		t.Errorf("expected vortex_print_i32 in IR:\n%s", ir)
	}
}

func TestCodegenStringLitGlobal(t *testing.T) {
	ir := parseAndCodegen(t, `let s: string = "hi";`)
	if !strings.Contains(ir, "c\"hi\\00\"") {
		t.Errorf("expected string constant c\"hi\\00\" in IR:\n%s", ir)
	}
}

func TestCodegenTensorLet(t *testing.T) {
	ir := parseAndCodegen(t, "let t: tensor<f32, [2, 3]> = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];")
	if !strings.Contains(ir, "call %VortexTensor* @vortex_tensor_create") {
		t.Errorf("expected call to vortex_tensor_create in IR:\n%s", ir)
	}
}

func TestCodegenTensorMatmul(t *testing.T) {
	ir := parseAndCodegen(t, `
		let matA: tensor<f32, [2, 3]> = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
		let matB: tensor<f32, [3, 2]> = [7.0, 8.0, 9.0, 10.0, 11.0, 12.0];
		let matC = matA * matB;
	`)
	if !strings.Contains(ir, "call void @vortex_matmul") {
		t.Errorf("expected call to vortex_matmul in IR:\n%s", ir)
	}
}

func TestCodegenTensorAdd(t *testing.T) {
	ir := parseAndCodegen(t, `
		let matA: tensor<f32, [2, 3]> = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
		let matB: tensor<f32, [2, 3]> = [7.0, 8.0, 9.0, 10.0, 11.0, 12.0];
		let matC = matA + matB;
	`)
	if !strings.Contains(ir, "call void @vortex_tensor_add") {
		t.Errorf("expected call to vortex_tensor_add in IR:\n%s", ir)
	}
}

func TestCodegenTensorRelu(t *testing.T) {
	ir := parseAndCodegen(t, `
		let mat: tensor<f32, [2, 3]> = [1.0, -2.0, 3.0, -4.0, 5.0, -6.0];
		let r = relu(mat);
	`)
	if !strings.Contains(ir, "call void @vortex_tensor_relu") {
		t.Errorf("expected call to vortex_tensor_relu in IR:\n%s", ir)
	}
}

func TestCodegenTensorSigmoid(t *testing.T) {
	ir := parseAndCodegen(t, `
		let mat: tensor<f32, [2, 3]> = [0.0, 1.0, -1.0, 2.0, -2.0, 0.5];
		let s = sigmoid(mat);
	`)
	if !strings.Contains(ir, "call void @vortex_tensor_sigmoid") {
		t.Errorf("expected call to vortex_tensor_sigmoid in IR:\n%s", ir)
	}
}

func TestCodegenTensorLetWithoutValue(t *testing.T) {
	ir := parseAndCodegen(t, "let t: tensor<f32, [2, 3]> = [0.0, 0.0, 0.0, 0.0, 0.0, 0.0];")
	if !strings.Contains(ir, "call %VortexTensor* @vortex_tensor_create") {
		t.Errorf("expected call to vortex_tensor_create in IR:\n%s", ir)
	}
}
