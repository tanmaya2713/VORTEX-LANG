# Vortex Architecture

This document describes the internal architecture of the Vortex compiler — how source code becomes a native binary and how the mobile strategy works end-to-end.

---

## Compilation Pipeline

```
 .vx source
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

The lexer scans the source text rune-by-rune and produces a slice of `common.Token` values. It recognizes:

- **47 keywords** via `src/dict/dictionary.go` — `fn`, `let`, `mut`, `if`, `else`, `for`, `while`, `return`, `break`, `continue`, `print`, `true`, `false`, `struct`, `model`, `layer`, `train`, `import`, `assert`, `tensor`, `ref`, `type`, `epochs`, `lr`, `strategy`, `devices`, type names, and AI operator keywords (`dense`, `conv2d`, `relu`, `sigmoid`, `adam`, etc.)
- **Types:** `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`, `f32`, `f64`, `bool`, `string`, `void`, `tensor<N,M>`
- **Operators:** `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `=`, `&&`, `||`, `!`, `&`, `|`, `^`, `~`, `@`, `#`, `->`
- **Symbols:** `;`, `:`, `,`, `.`, `(`, `)`, `{`, `}`, `[`, `]`
- **Comments:** Line comments (`//`) and nestable block comments (`/* */`)
- **Strings:** Double and single quoted with escape sequences (`\n`, `\t`, `\\`, `\"`, `\'`)
- **Noise filtering:** 141 English noise words (articles, prepositions, conjunctions) are removed after `let`/`mut` to enable natural-language syntax
- **Numbers:** Integers and floats with scientific (`e`/`E`) exponent notation

```go
// Output type (from src/common/types.go)
type Token struct {
    Kind   TokenKind
    Lexeme string
    Pos    Position
    Literal interface{}
}
```

### Stage 2 — Parser

**Package:** `src/parser/parser.go`

The parser is a recursive-descent parser with packrat-style memoization and precedence climbing. It consumes the token stream and builds an `*ast.Program` — a concrete syntax tree with 30+ node types. It handles:

- **Let bindings** with optional `mut` flag and type annotations (`let x: i32 = 10`)
- **Function definitions** with named parameters and return types
- **Function calls** with argument arity validation
- **If/else** conditionals with arbitrary nesting
- **While loops** with `break` / `continue`
- **For loops** (`for var in expr`) with index variable allocation
- **Struct definitions** with named fields
- **Model / layer definitions** (`model`, `layer`) with named parameters for AI architecture declarations
- **Train statements** (`train model, data, epochs, lr, strategy, devices`)
- **Print / assert / import** statements
- **Operator precedence** via precedence climbing (precedence table with left/right binding power)
- **Tensor literal parsing** (`[[...], [...]]`)
- **Array literals** (`[1, 2, 3]`)
- **Member expressions** (`obj.field`) and **index expressions** (`arr[i]`)
- **Assign expressions** (`x = expr`)
- **Type expressions** via `TypeExpr{Name, Args}` — handles reference types, explicit type annotations, and `ref` type structures

### Stage 3 — Type Checker

**Package:** `src/types/checker.go`

The type checker operates in **two passes** (`collectDecls` → `checkProgram`) with a lexical **`Scope`** chain (parent-linked for block scoping).

**Pass 1 — Declaration Collection:** Gathers all function signatures, struct definitions, and model declarations into a symbol table without processing bodies.

**Pass 2 — Full Checking:** Walks every statement and expression in the typed AST:

- Infers types for literal expressions (`42` → `i32`, `3.14` → `f64`, `"hi"` → `string`, `true` → `bool`)
- Validates type annotations against inferred types: `let x: i32 = "hello"` → compile error
- Validates condition types (must be `bool`): `if 42 { ... }` → error
- Checks function call arity and argument types against declared parameter types
- Resolves tensor dimensions (`tensor<N,M>`) and validates element type
- Checks array literal element type consistency
- Validates struct field access types via `MemberExpr`
- Enforces block scoping rules — variable lookup traverses the parent `Scope` chain
- Performs numeric type promotion (`i32` → `f64`) via `commonType`
- Validates `model`/`layer` declarations and `train` statement argument types
- Special-cases builtins `relu`/`sigmoid` for tensor argument validation

Type resolution happens **before** any code generation — no type errors reach LLVM.

### Stage 4 — LLVM IR Generation

**Package:** `src/codegen/llvmir/`

The codegen module walks the typed AST and produces LLVM IR as a string via the [`llvm`](https://github.com/llir/llvm) Go library.

```
LetStatement    → alloca + store (tensor path uses `vortex_tensor_create`)
BinaryExpr      → add/sub/mul/fdiv/icmp (dispatch int vs float via `llvm_types.go`)
IfExpr          → br/phi (SSA)
WhileLoop       → br/condbr/phi with break/continue stacks
ForLoop         → index alloca + header/body/exit blocks with element extraction alloca
FnDef           → define function with typed params and return
FnCall          → call instruction (builtins `relu`/`sigmoid` special-cased)
Return          → ret terminator (void or value)
Break/Continue  → branch to top of break/continue stack
Assign          → store to ident alloca or array index store-through-temp
Assert          → cond br to fail/continue with `printf` error message
Print           → type-dispatch to `vortex_print_i32`/`_f64`/`_bool`/`_string` + `_newline`
Tensor ops      → `vortex_tensor_create`, `vortex_matmul`, `vortex_tensor_add`, `vortex_tensor_relu`, `vortex_tensor_sigmoid`
Model/Train     → `vortex_init` call (train is a no-op stub)
```

All tensor operations (`matmul`, `relu`, `sigmoid`, `add`) are emitted as `call` instructions to the C runtime — LLVM optimizes and inlines them during compilation. The `%VortexTensor` struct type (`{ i32*, i32, float* }`) is injected into the module header automatically by `IRString()`.

### Stage 5 — Clang Compilation

The final step compiles the LLVM IR together with the embedded C runtime into a native binary.

```
clang out.ll io.c tensor.c -o <output>
```

**Cross-compilation** adds `-target <triple>`. See **🎯 Cross-Platform Compilation Targets** under Cross-Compilation Architecture below for the full target matrix.

---

## V8.1 Ascended Core — Python Interpreter

Alongside the native Go compiler pipeline, Vortex ships the **V8.1 Ascended Core** — a Python-based interpreter that provides OS-level I/O, Web API fetching, JSON-native parsing, Arrays, and an expanded standard library. This interpreter is bundled as `compiler.py` in every release and installed via the `vx` wrapper command.

### Architecture Overview

```
 .vx source
      │
      ▼
 ┌─────────────┐
 │   Lexer     │  Regex-based tokenization
 │ (compiler.py│  12 named capture groups
 │  lines 33–66)│  skips whitespace/comments
 └──────┬──────┘
        │ []Token tuples
        ▼
 ┌──────────────┐
 │  VortexEngine│  Two-pass evaluation
 │  (compiler.py│  Pass 1: function call interception
 │  lines 95–245)│  Pass 2: Python expression bridge
 └───────┬──────┘
         │ eval(expr_str, {"__mem__": memory})
         ▼
    Python Runtime
```

### Stage 1 — Regex Lexer

Unlike the Go lexer which uses a character-by-character state machine, the Python interpreter uses a single combined regex with named capture groups:

| Token Group | Pattern |
|-------------|---------|
| `COMMENT` | `//.*` |
| `NUMBER` | `\d+(\.\d+)?` |
| `STRING` | `".*?"\|'.*?'` |
| `KEYWORD` | `fn\|let\|if\|else\|while\|true\|false\|print\|return\|save\|load\|into\|request` |
| `BUILTIN` | `random\|sqrt\|round\|abs\|pow\|sin\|cos\|tan\|length\|type_of\|str` |
| `IDENTIFIER` | `[a-zA-Z_][a-zA-Z0-9_]*` |
| `OPERATOR` | `==\|!=\|>=\|<=\|>\|<\|=\|+\|-\|*\|\|\|&&\|,` |
| `PUNCT` | `[\[\]{}();:]` |

The lexer produces `(kind, value)` tuples, filtering out `SPACE` and `COMMENT` groups.

### Stage 2 — Memory Stack (Isolated RAM)

Each function call gets its own isolated `VortexMemoryStack`:

```
 ┌─────────────────────┐
 │  Global Memory      │  Root scope
 │  {x: 10, y: "hi"}   │
 └─────────────────────┘
         │
         ▼
 ┌─────────────────────┐
 │  Function Frame     │  Copied from global, then locals allocated
 │  {x: 10, ...,       │
 │   param: val}       │  Parameters bound here
 └─────────────────────┘
```

The `VortexMemoryStack` class provides `allocate()`, `update()`, `fetch()`, and `dump_core()` methods. Each inner `VortexEngine` receives a copy-on-write snapshot of the parent's memory, ensuring function purity and preventing side effects.

### Stage 3 — Evaluation Engine

The `evaluate()` method is a **two-pass** system:

**Pass 1 — Function Interception:** Before any math evaluation, the engine scans for user-defined function calls (`fn`), extracts arguments, evaluates them recursively, spins up an isolated `VortexEngine` with its own `MemoryStack`, executes the function body, captures the `return_value`, and substitutes the result back into the expression.

**Pass 2 — Python Expression Bridge:** The engine builds a Python expression string from the resolved tokens:
- Identifiers → `__mem__["var_name"]` lookups
- Builtins → direct Python function calls (`math.sqrt()`, `random.uniform()`, etc.)
- Operators → Python operators (with `&&` → `and`, `||` → `or`)
- Booleans → Python `True`/`False`

The assembled string is executed via:
```python
eval(expr_str, {"__mem__": self.stack.dump_core()})
```

This gives Vortex scripts full arithmetic, boolean logic, and string concatenation without building a custom expression parser. Security is maintained by restricting the globals dictionary to only `__mem__`.

### JavaScript-Style String Coercion

If `eval()` fails and the expression contains `+` operators, the engine falls back to a string concatenation mode — splitting on `+`, recursively evaluating each sub-expression, and concatenating the results as strings. This mirrors JavaScript's dynamic type coercion:

```vortex
let result = "The answer is " + 42; // → "The answer is 42"
```

### Omniversal Capabilities

#### OS File I/O (`save` / `load`)

The `save` keyword writes any variable to the filesystem. If the value is a dictionary or list, it is serialized as JSON. The `load` keyword reads a file and auto-detects the format: JSON is parsed into native dicts/lists, numbers are converted to int/float, and everything else is stored as a string.

```vortex
save my_data into "output.json";
load "config.json" into config;
```

#### Web API Fetching (`request`)

The `request` keyword performs a read-only HTTP GET request and stores the response. JSON responses are automatically parsed into native data structures; plain text responses are stored as strings.

```vortex
request "https://api.example.com/data" into response;
print(response["status"]);
```

All `request` calls include a `User-Agent: Vortex-Engine/1.0` header. Network errors are caught and stored as descriptive error strings rather than crashing the runtime.

---

## C Runtime Embedding

**File:** `src/runtime/embed.go`

```go
package runtime

import "embed"

//go:embed c_lib/io.c c_lib/tensor.c c_lib/tensor.h
var RuntimeFS embed.FS
```

The three C source files under `src/runtime/c_lib/` are embedded into the Go binary at compile time via Go 1.16's `//go:embed` directive.

### During compilation (`compileToBinary` in `cmd/vortex/main.go`):

1. A temporary directory is created via `os.MkdirTemp`
2. `out.ll` is written from the generated LLVM IR
3. `io.c`, `tensor.c`, `tensor.h` are extracted from `RuntimeFS` into the same temp directory
4. Clang is invoked to compile the LLVM IR together with the C files

This produces a **statically linked binary** — no shared library dependencies at runtime.

### C Runtime Files

| File | Purpose |
|------|---------|
| `io.c` | `vortex_print_i32(i32)`, `vortex_print_f64(f64)`, `vortex_print_bool(bool)`, `vortex_print_string(str)`, `vortex_print_newline()` — system I/O bindings |
| `tensor.c` | `vortex_tensor_create`, `vortex_matmul`, `vortex_tensor_add`, `vortex_tensor_relu`, `vortex_tensor_sigmoid`, `vortex_print_tensor` — tensor math operations with heap-allocated buffers |
| `tensor.h` | `VortexTensor` struct type definition (`int* dims`, `int ndim`, `float* data`) and function declarations shared between generated code and C runtime |

---

## Mobile Strategy

### Approach 1: Termux Native CLI

[Termux](https://termux.com/) provides a Linux environment on Android. The `--target android` flag produces an ELF binary that runs directly in Termux.

```
vortex build app.vx --target android
adb push app /data/local/tmp/
adb shell /data/local/tmp/app
```

### Approach 2: Android Shared Library (JNI)

The `--shared` flag produces a `.so` file loadable by Android's `System.loadLibrary()`.

```
vortex build engine.vx --target android --shared
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

### 🎯 Cross-Platform Compilation Targets

| Target OS | Architecture | Clang Target Triple | Deployment Strategy / Notes |
| :--- | :--- | :--- | :--- |
| **Windows** | `x86_64` (AMD64) | `x86_64-pc-windows-msvc` | Core desktop execution target. |
| **Linux** | `x86_64` (AMD64) | `x86_64-unknown-linux-gnu` | Standard cloud/server environment compatibility. |
| **Linux** | `aarch64` (ARM64) | `aarch64-unknown-linux-gnu` | Edge computing and modern ARM64 server support. |
| **macOS** | `arm64` (Apple Silicon) | `arm64-apple-darwin` | Apple Silicon native exclusive. Intel support deprecated. |
| **Android** | `aarch64` (ARM64) | `aarch64-linux-android` | Android native runtime integration. |

> **System Note:** Compiling the engine in Shared Library mode via the `--shared` flag forces the absolute application of `-shared -fPIC` compilation flags, standardizing the native binary output layout to `lib<name>.so` for smooth foreign function interface (FFI) execution.

---

## Directory Layout

```
cmd/vortex/main.go                  CLI entry point (flags, REPL, build/run/compile)
compiler.py                         V8.1 Ascended Core Python interpreter
.github/workflows/release.yml       CI/CD pipeline (6-matrix cross-compilation)
dist/install.sh / install.ps1       Platform installers (POSIX + Windows)
assets/vortex-logo.png              Brand assets
docs/                               Documentation index (placeholder)
examples/                           Example .vx programs
scripts/                            Build and utility scripts
tests/main.vx                       E2E integration stress test benchmark
.gitignore                          Git ignore rules (vortex.exe, __pycache__)
go.mod                              Go module definition
go.sum                              Go module checksum
LICENSE                             License file
CODE_OF_CONDUCT.md                  Contributor Covenant v2.1
README.md                           Project documentation
SECURITY.md                         Security policy
src/
├── ast/                            AST node definitions and pretty-printer
│   ├── ast.go                      All node types (30+ — Stmt, Expr, Program)
│   └── printer.go                  S-expression pretty-printer
├── codegen/
│   ├── cpu/                        CPU code generation target (reserved)
│   ├── edge/                       Edge code generation target (reserved)
│   ├── gpu/                        GPU code generation target (reserved)
│   └── llvmir/                     LLVM IR code generation
│       ├── codegen.go              Module/function compilation orchestrator
│       ├── codegen_test.go         Codegen test suite
│       ├── expr.go                 Expression codegen (659 lines)
│       ├── stmt.go                 Statement codegen (404 lines)
│       ├── llvm_types.go           Type mapping (Vortex → LLVM)
│       └── runtime.go              C runtime function declarations
├── common/
│   └── types.go                    Shared types (Position, Token, TokenKind, Symbol, DataType)
├── dict/
│   └── dictionary.go               Keyword dictionary (47 keywords), noise words (141), ML operators
├── layout/                         Layout analysis (reserved)
├── lexer/                          Rune-based lexer with noise filtering
│   ├── lexer.go                    Lexer implementation (313 lines)
│   └── lexer_test.go               Lexer test suite
├── mlir/                           MLIR dialect integration (reserved)
├── parser/                         Recursive-descent parser with packrat memoization
│   ├── parser.go                   Parser implementation (990 lines)
│   └── parser_test.go              Parser test suite
├── runtime/
│   ├── c_lib/                      Embedded C runtime source files
│   │   ├── io.c                    Print I/O bindings (i32, f64, bool, string, newline)
│   │   ├── tensor.c                Tensor math (matmul, add, relu, sigmoid, create, print)
│   │   └── tensor.h                VortexTensor struct type and function declarations
│   └── embed.go                    //go:embed directive for C runtime extraction
├── safety/                         Safety and memory analysis (reserved)
├── types/                          Type checker
│   ├── checker.go                  Two-pass type checker (739 lines)
│   ├── checker_test.go             Type checker test suite (258 lines)
│   └── type.go                     Type system (Primitive, FnType, StructType, ArrayType, TensorType, etc.)
└── integration_test.go             End-to-end pipeline tests (lex → parse → typecheck → codegen → clang → run)
```

---

## Support

For architectural questions or technical inquiries, contact the engineering team at **[dronalabs.support@gmail.com](mailto:dronalabs.support@gmail.com)**.
