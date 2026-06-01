package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/codegen/llvmir"
	"github.com/vortex-lang/vortex/src/common"
	"github.com/vortex-lang/vortex/src/lexer"
	"github.com/vortex-lang/vortex/src/parser"
	vxruntime "github.com/vortex-lang/vortex/src/runtime"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

type Compiler struct {
	file    string
	source  string
	tokens  []common.Token
	program *ast.Program
	errs    []*common.CompilerError
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) CompileFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	c.file = path
	c.source = string(data)
	return c.Compile()
}

func (c *Compiler) CompileSource(source string, name string) error {
	c.file = name
	c.source = source
	return c.Compile()
}

func (c *Compiler) Compile() error {
	tokens, err := lexer.Lex(c.source, c.file)
	if err != nil {
		return err
	}
	tokens = lexer.FilterNoise(tokens)
	c.tokens = tokens

	prog, errs := parser.ParseTokens(tokens, c.file)
	c.program = prog
	c.errs = errs

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (c *Compiler) DumpTokens() {
	for _, tok := range c.tokens {
		fmt.Println(tok)
	}
}

func (c *Compiler) DumpAST() {
	if c.program != nil {
		printer := ast.NewPrinter()
		printer.Print(c.program)
	}
}

func runRepl() {
	fmt.Println("Vortex v0.1.0 — type :quit to exit")
	compiler := NewCompiler()
	for {
		fmt.Print("vtx> ")
		var input string
		fmt.Scanln(&input)
		if input == ":quit" || input == ":q" {
			break
		}
		if input == "" {
			continue
		}
		err := compiler.CompileSource(input, "<repl>")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		compiler.DumpAST()
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			cmdBuild(os.Args[2:])
			return
		case "run":
			cmdRun(os.Args[2:])
			return
		}
	}

	repl := flag.Bool("repl", false, "Start interactive REPL")
	tokens := flag.Bool("tokens", false, "Dump tokens")
	ast := flag.Bool("ast", false, "Dump AST")
	emit := flag.String("emit", "", "Emit stage (llvm)")
	output := flag.String("o", "", "Output file")
	flag.Parse()

	if *repl {
		runRepl()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: vortex [--repl] [--tokens] [--ast] [--emit=llvm] [-o output] <file.vtx>\n")
		os.Exit(1)
	}

	path := args[0]
	compiler := NewCompiler()

	err := compiler.CompileFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *tokens {
		compiler.DumpTokens()
	}
	if *ast {
		compiler.DumpAST()
	}

	if *emit == "llvm" {
		checker := vtypes.New()
		ok := checker.Check(compiler.program)
		if !ok {
			errs := checker.Errors()
			if len(errs) > 0 {
				fmt.Fprintf(os.Stderr, "Type error: %v\n", errs[0])
			} else {
				fmt.Fprintf(os.Stderr, "Type checking failed\n")
			}
			os.Exit(1)
		}

		cg := llvmir.New()
		mod := cg.Compile(compiler.program)
		llvmIR := mod.String()

		if *output != "" {
			outPath := *output
			dir := filepath.Dir(outPath)
			if dir != "." {
				os.MkdirAll(dir, 0755)
			}
			err := os.WriteFile(outPath, []byte(llvmIR), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Write output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("LLVM IR written to %s\n", outPath)
		} else {
			fmt.Print(llvmIR)
		}
		return
	}

	if *output != "" {
		outPath := *output
		dir := filepath.Dir(outPath)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		err := os.WriteFile(outPath, []byte("; Vortex compiled output\n"), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", outPath)
	}
}

func compileToBinary(inputFile, binPath, targetOS string, shared bool) error {
	compiler := NewCompiler()
	if err := compiler.CompileFile(inputFile); err != nil {
		return err
	}
	checker := vtypes.New()
	if !checker.Check(compiler.program) {
		errs := checker.Errors()
		if len(errs) > 0 {
			return errs[0]
		}
		return fmt.Errorf("type checking failed")
	}
	cg := llvmir.New()
	cg.Compile(compiler.program)
	if len(cg.Errors()) > 0 {
		return cg.Errors()[0]
	}
	tmpDir, err := os.MkdirTemp("", "vortex-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	llPath := filepath.Join(tmpDir, "out.ll")
	if err := os.WriteFile(llPath, []byte(cg.IRString()), 0644); err != nil {
		return fmt.Errorf("write .ll: %w", err)
	}
	files := map[string]string{
		"io.c":     "c_lib/io.c",
		"tensor.c": "c_lib/tensor.c",
		"tensor.h": "c_lib/tensor.h",
	}
	for outName, srcPath := range files {
		src, err := vxruntime.RuntimeFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", srcPath, err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, outName), src, 0644); err != nil {
			return fmt.Errorf("write %s: %w", outName, err)
		}
	}
	if _, err := exec.LookPath("clang"); err != nil {
		return fmt.Errorf("Clang not found in PATH. vortex build requires Clang.")
	}
	clangArgs := []string{"out.ll", "io.c", "tensor.c", "-I" + tmpDir, "-lm", "-o", binPath}
	if shared {
		clangArgs = append(clangArgs, "-shared", "-fPIC")
	}
	switch targetOS {
	case "windows":
		clangArgs = append(clangArgs, "-target", "x86_64-pc-windows-msvc")
	case "linux":
		clangArgs = append(clangArgs, "-target", "x86_64-unknown-linux-gnu")
	case "macos":
		if runtime.GOARCH == "arm64" {
			clangArgs = append(clangArgs, "-target", "arm64-apple-darwin")
		} else {
			clangArgs = append(clangArgs, "-target", "x86_64-apple-darwin")
		}
	case "android":
		clangArgs = append(clangArgs, "-target", "aarch64-linux-android")
	}
	cmd := exec.Command("clang", clangArgs...)
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clang failed: %w\n%s", err, out)
	}
	return nil
}

func getDefaultTarget() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	default:
		return "linux"
	}
}

func cmdBuild(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: vortex build <file.vtx> [-o <output>] [--target <target>] [--shared]\n")
		os.Exit(1)
	}
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
	output := buildFlags.String("o", "", "Output binary path")
	target := buildFlags.String("target", "", "Target platform (windows, linux, macos, android)")
	shared := buildFlags.Bool("shared", false, "Build as shared library (.so)")
	buildFlags.Parse(args)
	inputFile := buildFlags.Arg(0)
	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: vortex build <file.vtx> [-o <output>] [--target <target>] [--shared]\n")
		os.Exit(1)
	}
	targetOS := *target
	if targetOS == "" {
		targetOS = getDefaultTarget()
	}
	outputPath := *output
	if outputPath == "" {
		base := filepath.Base(inputFile)
		base = strings.TrimSuffix(base, ".vtx")
		switch {
		case *shared:
			base = "lib" + base + ".so"
		case targetOS == "windows":
			base += ".exe"
		}
		outputPath = base
	}
	outputPath, err := filepath.Abs(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving output path: %v\n", err)
		os.Exit(1)
	}
	if err := compileToBinary(inputFile, outputPath, targetOS, *shared); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Built %s -> %s\n", inputFile, outputPath)
}

func cmdRun(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: vortex run <file.vtx>\n")
		os.Exit(1)
	}
	inputFile := args[0]
	tmpDir, err := os.MkdirTemp("", "vortex-run-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	binName := "vortex_prog"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tmpDir, binName)
	if err := compileToBinary(inputFile, binPath, getDefaultTarget(), false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command(binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running binary: %v\n", err)
		os.Exit(1)
	}
}
