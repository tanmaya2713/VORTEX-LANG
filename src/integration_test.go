package src

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vortex-lang/vortex/src/codegen/llvmir"
	"github.com/vortex-lang/vortex/src/lexer"
	"github.com/vortex-lang/vortex/src/parser"
	vxruntime "github.com/vortex-lang/vortex/src/runtime"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

func TestClangEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not found in PATH; skipping integration test")
	}

	runtimeDir := t.TempDir()
	files := map[string]string{
		"io.c":     "c_lib/io.c",
		"tensor.c": "c_lib/tensor.c",
		"tensor.h": "c_lib/tensor.h",
	}
	for outName, srcPath := range files {
		src, err := vxruntime.RuntimeFS.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read embedded %s: %v", srcPath, err)
		}
		if err := os.WriteFile(filepath.Join(runtimeDir, outName), src, 0644); err != nil {
			t.Fatalf("write %s: %v", outName, err)
		}
	}

	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"print_i32", "print 42;", "42\n"},
		{"print_f64", "print 3.14;", "3.14\n"},
		{"print_bool_true", "print true;", "true\n"},
		{"print_bool_false", "print false;", "false\n"},
		{"print_string", `print "hello world";`, "hello world\n"},
		{"print_ident", "let x = 10; print x;", "10\n"},
		{"print_add", "print 1 + 2;", "3\n"},
		{"print_multi_stmt", "let x = 5; let y = 3; print x + y;", "8\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := lexer.Lex(tt.source, "test.vtx")
			if err != nil {
				t.Fatalf("lex: %v", err)
			}
			tokens = lexer.FilterNoise(tokens)
			p := parser.New(tokens, "test.vtx")
			prog := p.Parse()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse: %v", p.Errors()[0])
			}
			checker := vtypes.New()
			if !checker.Check(prog) {
				t.Fatalf("typecheck: %v", checker.Errors()[0])
			}
			cg := llvmir.New()
			cg.Compile(prog)
			if len(cg.Errors()) > 0 {
				t.Fatalf("codegen: %v", cg.Errors()[0])
			}
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, "out.ll")
			if err := os.WriteFile(llFile, []byte(cg.IRString()), 0644); err != nil {
				t.Fatalf("write .ll: %v", err)
			}
			binPath := filepath.Join(tmpDir, "test_bin")
			cmd := exec.Command("clang", llFile, filepath.Join(runtimeDir, "io.c"), filepath.Join(runtimeDir, "tensor.c"), "-I"+runtimeDir, "-lm", "-o", binPath)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("clang failed: %v\noutput: %s", err, out)
			}
			execPath := binPath
			if runtime.GOOS == "windows" {
				if _, err := os.Stat(binPath); os.IsNotExist(err) {
					execPath = binPath + ".exe"
				}
			}
			cmd2 := exec.Command(execPath)
			out2, err := cmd2.CombinedOutput()
			if err != nil {
				t.Fatalf("exec failed: %v\noutput: %s", err, out2)
			}
			if string(out2) != tt.expected {
				t.Errorf("got %q; want %q", string(out2), tt.expected)
			}
		})
	}
}
