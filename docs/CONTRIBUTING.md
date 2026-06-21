# Contributing to the Vortex Engine

First off, thank you for considering contributing to Vortex! The V8.1 Ascended Core is an ambitious, LLVM-backed systems programming language, and community contributions are what will drive it to the next level.

This document outlines the process for contributing to the Drona Labs Vortex ecosystem.

## 🧠 Architectural Philosophy
Before diving into the code, please familiarize yourself with our pipeline:
1. **Frontend:** Regex/Packrat Lexer -> Recursive-Descent Parser -> Two-Pass Type Checker.
2. **Backend:** LLVM IR Codegen -> Native Clang Compilation + C Runtime linking.
3. **Omniversal:** Our Python interpreter (`compiler.py`) handles OS/Network abstractions.

Maintain strict parity between the Go compiler, the C runtime, and the documentation. 

## 🐛 Reporting Bugs
If you find a bug in the compiler or runtime, please open an issue containing:
* A clear, descriptive title.
* The exact `.vx` code snippet that caused the failure.
* The expected behavior vs. the actual output (including LLVM verifier errors or panics).
* Your OS and Clang version.

## ✨ Suggesting Enhancements
Want to add a new AST node, operator, or standard library feature?
* Open an issue proposing the feature.
* Detail the syntax, the type-checking rules, and how the LLVM IR generation should behave.
* Wait for a core maintainer to approve the architecture before writing the code.

## 🛠️ Pull Request Process
1. **Fork the repo** and create your branch from `main`.
2. **Write tests:** If you add a new AST node or feature, add a corresponding test in `checker_test.go` or `codegen_test.go`.
3. **Verify Parity:** Ensure your changes do not break the 100% parity between the codebase and `ARCHITECTURE.md` / `README.md`.
4. **Pass the Pipeline:** Run `go test ./...` and ensure all tests pass cleanly.
5. **Commit cleanly:** Use Conventional Commits (e.g., `feat(parser): add bitwise operators` or `fix(codegen): resolve segfault in matmul`).

## ⚖️ Code of Conduct
By participating in this project, you agree to abide by our `CODE_OF_CONDUCT.md`. We maintain a highly disciplined, respectful, and focused engineering environment.