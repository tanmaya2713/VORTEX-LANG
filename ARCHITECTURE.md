# Vortex Architecture

This document describes the internal architecture of the Vortex compiler — how source code becomes a native binary and how the mobile strategy works end-to-end.

---

## Compilation Pipeline

```
 .vtx source
      │
      ▼
 ┌─────────────┐
 │   Lexer     │  Go (src/lexer/)
 │  (lexer.go) │
 └──────┬──────┘
        │ []common.Token
        ▼
 ┌─────────────┐
 │   Parser    │  Go (src/parser/)
 │ (parser.go) │
 └──────┬──────┘
        │ *ast.Program
        ▼
 ┌──────────────┐
 │ Typechecker  │  Go (src/types/)
 │ (checker.go) │
 └──────┬───────┘
        │ typed *ast.Program
        ▼
 ┌───────────────┐
 │  LLVM IR Gen  │  Go (src/codegen/llvmir/)
 │  (codegen.go) │
 └───────┬───────┘
         │ LLVM IR (string)
         ▼
 ┌───────────────┐
 │  Clang        │  C (external binary)
 │  (clang)      │  + embedded C runtime
 └───────┬───────┘
         │ .exe / .so / Mach-O / ELF
         ▼
    Native Binary
```

### Stage 1 — Lexer

**Package:** `src/lexer/lexer.go`

The lexer scans the source text and produces a slice of `common.Token` values. It recognizes:

- Keywords: `fn`, `let`, `if`, `else`, `while`, `return`, `print`, `true`, `false`
- Types: `i32`, `f64`, `bool`, `string`, `tensor<N,M>`
- Operators: `+`, `-`, `*`, `/`, `==`, `<`, `>`, `!`
- Noise filtering removes comments and whitespace

```go
// Output type (from src/common/types.go)
type Token struct {
    Type  TokenType
    Value string
    Line  int
    Col   int
    File  string
}
```

### Stage 2 — Parser

**Package:** `src/parser/parser.go`

The parser consumes the token stream and builds an `*ast.Program` — a concrete syntax tree. It handles:

- Let bindings (with and without type annotations)
- Function definitions and calls
- If/else conditionals
- While loops
- Model definitions (AI layer declarations)
- Operator precedence via Pratt parsing
- Tensor literal parsing (`[[...], [...]]`)

### Stage 3 — Type Checker

**Package:** `src/types/checker.go`

The type checker walks the AST and:

- Infers types for literal expressions (e.g., `42` → `i32`, `3.14` → `f64`)
- Validates type annotations against inferred types
- Catches mismatches: `let x: i32 = "hello"` → compile error
- Validates condition types (must be `bool`)
- Checks function call arity and argument types
- Resolves tensor dimensions for `tensor<N,M>` declarations
- Enforces block scoping rules

Type resolution happens before any code generation — no type errors reach LLVM.

### Stage 4 — LLVM IR Generation

**Package:** `src/codegen/llvmir/`

The codegen module walks the typed AST and produces LLVM IR as a string via the [`llvm`](https://github.com/llir/llvm) Go library.

```
LetStatement   → alloca + store
BinaryExpr     → add/sub/mul/fdiv/icmp
IfExpr         → br/phi (SSA)
WhileLoop      → br/condbr/phi
FnDef          → define function
FnCall         → call instruction
Tensor ops     → calls to embedded C functions
Print          → calls to vortex_print / vortex_print_str
```

All tensor operations (`matmul`, `relu`, `sigmoid`, `add`) are emitted as `call` instructions to the C runtime — LLVM optimizes and inlines them during compilation.

### Stage 5 — Clang Compilation

The final step compiles the LLVM IR together with the embedded C runtime into a native binary.

```
clang out.ll io.c tensor.c -o <output>
```

**Cross-compilation** adds `-target <triple>`:

| `--target` | Clang Triple |
|------------|-------------|
| `windows` | `x86_64-pc-windows-msvc` |
| `linux` | `x86_64-unknown-linux-gnu` |
| `macos` | `arm64-apple-darwin` or `x86_64-apple-darwin` |
| `android` | `aarch64-linux-android` |

**Shared library mode** (`--shared`) adds `-shared -fPIC` and renames output to `lib<name>.so`.

---

## C Runtime Embedding

**File:** `src/runtime/embed.go`

```go
package runtime

import "embed"

//go:embed io.c tensor.c tensor.h
var RuntimeFS embed.FS
```

The three C source files are embedded into the Go binary at compile time via Go 1.16's `//go:embed` directive.

### During compilation (`compileToBinary` in `cmd/vortex/main.go`):

1. A temporary directory is created via `os.MkdirTemp`
2. `out.ll` is written from the generated LLVM IR
3. `io.c`, `tensor.c`, `tensor.h` are extracted from `RuntimeFS` into the same temp directory
4. Clang is invoked to compile the LLVM IR together with the C files

This produces a **statically linked binary** — no shared library dependencies at runtime.

### C Runtime Files

| File | Purpose |
|------|---------|
| `io.c` | `vortex_print(i32)`, `vortex_print_f64(f64)`, `vortex_print_str(str)`, `vortex_print_tensor(...)` — system I/O bindings |
| `tensor.c` | `vortex_matmul`, `vortex_relu`, `vortex_sigmoid`, `vortex_add` — tensor math operations with heap-allocated buffers |
| `tensor.h` | Type definitions and function declarations shared between generated code and C runtime |

---

## Mobile Strategy

### Approach 1: Termux Native CLI

[Termux](https://termux.com/) provides a Linux environment on Android. The `--target android` flag produces an ELF binary that runs directly in Termux.

```
vortex build app.vtx --target android
adb push app /data/local/tmp/
adb shell /data/local/tmp/app
```

### Approach 2: Android Shared Library (JNI)

The `--shared` flag produces a `.so` file loadable by Android's `System.loadLibrary()`.

```
vortex build engine.vtx --target android --shared
# outputs: libengine.so
```

**Integration with Android Studio:**

```
app/src/main/jniLibs/arm64-v8a/libengine.so
```

```kotlin
class InferenceEngine {
    init {
        System.loadLibrary("engine")
    }

    // Native methods defined in the Vortex source
    external fun forward(input: FloatArray): FloatArray
}
```

The C runtime is compiled with `-fPIC` (position-independent code) and linked directly into the `.so`, so no additional native dependencies need to be shipped.

---

## Cross-Compilation Architecture

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────────┐
│  Host    │    │  Vortex  │    │  LLVM IR │    │  Clang    │
│ (any OS) │───▶│ Compiler │───▶│  (text)  │───▶│ -target   │───▶ Native Binary
└──────────┘    └──────────┘    └──────────┘    │ <triple>  │
                                                 └───────────┘
```

The Vortex compiler itself is written in Go and runs on any platform. The LLVM IR it generates is platform-agnostic. Clang (installed separately) performs the actual cross-compilation using the appropriate target triple. This means:

- **One compiler binary** — no need for separate cross-compiler builds of Vortex
- **Clang handles the heavy lifting** — target-specific codegen, linking, and optimization
- **CI-friendly** — GitHub Actions matrix builds using `GOOS`/`GOARCH` for the Vortex binary itself, Clang for the compiled output

---

## Directory Layout

```
cmd/vortex/main.go         CLI entry point
src/
├── ast/                   AST node definitions and printer
├── codegen/llvmir/        LLVM IR code generation
│   ├── codegen.go         Module and function compilation
│   ├── expr.go            Expression codegen
│   ├── stmt.go            Statement codegen
│   ├── llvm_types.go      Type mapping (Vortex → LLVM)
│   └── runtime.go         C runtime call generation
├── common/                Shared token and error types
├── dict/                  Dictionary data structure
├── lexer/                 Lexer / tokenizer
├── parser/                Recursive descent parser
├── runtime/               Embedded C source files
└── types/                 Type checker
    ├── checker.go         Type inference and validation
    ├── type.go            Type definitions
    └── checker_test.go    Type system tests
```

---

## Support

For architectural questions or technical inquiries, contact the engineering team at **[dronalabs.support@gmail.com](mailto:dronalabs.support@gmail.com)**.
