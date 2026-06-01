# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| v0.1.x  | ✅ Active development — security patches within 48 hours |
| < 0.1   | ❌ Pre-release, no guarantee |

---

## Reporting a Vulnerability

We take security seriously. If you discover a vulnerability in Vortex, report it privately before public disclosure.

**Do not open a public GitHub issue.**

### Contact

Send a detailed report to **[dronalabs.support@gmail.com](mailto:dronalabs.support@gmail.com)**. Do not open public GitHub issues for security vulnerabilities.

### Response Timeline

| Event | Target |
|-------|--------|
| Acknowledgment | 48 hours |
| Initial triage | 5 business days |
| Fix released | 90 days (coordinated disclosure) |
| Public advisory | After fix is tagged |

We commit to coordinated disclosure — you will be credited in the advisory unless you request anonymity.

---

## Security by Design

Vortex is built with several architectural properties that reduce the surface area for vulnerabilities:

### 1. LLVM as a Security Boundary

Vortex delegates native code generation and optimization to [LLVM](https://llvm.org/). This provides:

- **Compiler-verified safety** — LLVM's `verifyModule()` pass catches undefined behavior, mismatched types, and invalid control flow before machine code is emitted
- **Static analysis** — LLVM's optimizer eliminates dead code, detects unreachable paths, and enforces data-flow consistency
- **Proven correctness** — LLVM has 20+ years of production use across C/C++/Rust compilers; the Vortex frontend inherits this reliability

The Vortex frontend never emits raw machine code — only LLVM IR, which passes through LLVM's verifier before compilation.

### 2. Go Frontend Type Safety

The entire frontend (lexer, parser, type checker, IR generator) is written in **Go**, a memory-safe language with garbage collection and no unsafe pointer arithmetic in the Vortex codebase.

#### Type System Guards

| Check | Example | Protection |
|-------|---------|------------|
| Type inference | `let x = 42` → `i32` | Prevents implicit type confusion |
| Annotation validation | `let x: i32 = "hello"` → error | Catches logic errors at compile time |
| Condition typing | `if 42 { ... }` → error (not bool) | Prevents non-boolean condition bugs |
| Function arity | `fn f(a, b)` called as `f(1)` → error | Prevents stack corruption from arg mismatch |
| Tensor dimension | `tensor<2,3>` vs actual shape | Prevents buffer over-reads in C runtime |
| Block scoping | Variable access outside scope → error | Prevents use-after-free patterns |

All type errors are caught **before** LLVM IR generation — no malformed IR can reach Clang.

### 3. Go Runtime Safety

Vortex uses Go 1.16+ `//go:embed` to ship the C runtime:

```go
//go:embed io.c tensor.c tensor.h
var RuntimeFS embed.FS
```

This is a **read-only embed at compile time** — no file system access at runtime, no path traversal vectors, no mutable global state in the embed layer.

---

## C Runtime Memory Safety

The boundary between the Go frontend and the embedded C runtime is the critical security perimeter.

### Architecture of the Boundary

```
┌──────────────────────────────────────────────┐
│           Vortex Frontend (Go)               │
│  Memory-safe, garbage-collected, typed       │
│  Generates LLVM IR with typed call           │
│  instructions to C runtime functions         │
└────────────┬─────────────────────────────────┘
             │ call vortex_matmul(a, b, result)
             │
             ▼
┌──────────────────────────────────────────────┐
│           C Runtime (io.c, tensor.c)          │
│  Compiled with Clang -fstack-protector       │
│  Bounded loops, validated dimensions         │
│  No inline assembly                          │
└──────────────────────────────────────────────┘
```

### Runtime Defenses

| Measure | Implementation |
|---------|----------------|
| Stack canaries | C code compiled with `-fstack-protector-strong` |
| Dimension validation | Tensor dimensions are validated by the Go typechecker before any C function is called; the C functions trust but verify |
| Heap bounds | All tensor operations use pre-computed sizes; loops iterate over validated dimension counts |
| No inline assembly | The C runtime is pure C — no `asm()` blocks, no architecture-specific exploits |
| No external dependencies | The C runtime links only against libc (via Clang) — no third-party native libraries |

### What the C Runtime Does NOT Do

- **No file I/O** — file operations are handled by the Go CLI, never by generated code
- **No network access** — the C runtime has no socket or network functionality
- **No dynamic allocation exposed** — memory allocation is scoped to tensor operations and freed before returning to Vortex code
- **No user input parsing** — strings are printed, not executed or parsed

---

## Supply Chain Security

### Build Pipeline

- **GitHub Actions CI** — every push runs `go vet ./...` and the full test suite
- **Matrix builds** — release artifacts are built on GitHub's infrastructure, not developer machines
- **Dependency verification** — `go.sum` locks all Go module hashes; `go mod verify` checks integrity
- **No prebuilt binaries** — all release binaries are built from source in CI

### Recommended Verification

```bash
# Verify module integrity
go mod verify

# Vet the codebase
go vet ./...

# Run the test suite
go test ./...
```

---

## Security FAQ

**Q: Can Vortex code cause a buffer overflow in the C runtime?**

Only if a malicious actor bypasses the Go typechecker and emits raw LLVM IR. The standard compilation pipeline validates all tensor dimensions before codegen. The C runtime functions also perform internal bounds checks on the passed dimensions as a defense-in-depth measure.

**Q: Is the `--shared` output safe to load in Android?**

Yes. The `.so` produced by `--target android --shared` is compiled with `-fPIC` and `-fstack-protector-strong`. All Vortex type checking applies before the shared library is built. The resulting `.so` has no external native dependencies beyond standard libc.

**Q: What if Clang has a vulnerability?**

Clang is a separate binary in `$PATH` — Vortex invokes it as a subprocess. A Clang vulnerability would be a supply-chain issue, not a Vortex issue. We recommend keeping Clang updated via your system package manager. Our CI uses the latest stable LLVM release.

**Q: Does Vortex use `unsafe` Go code?**

No. The Vortex codebase contains zero `unsafe` imports. All frontend memory management is handled by Go's garbage collector.

---

## Drona Labs Security Sign-off

Vortex is developed and maintained by **Drona Labs** under the engineering leadership of **Tanmaya Mahapatra (Retired Ame)**. We are committed to delivering a secure, robust, and auditable programming language for the AI and systems programming community.

For security-critical inquiries, reach us at **[dronalabs.support@gmail.com](mailto:dronalabs.support@gmail.com)**.
