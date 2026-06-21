<div align="center">
  <img src="https://github.com/tanmaya2713/VORTEX-LANG/blob/main/assets/vortex-logo.png?raw=true" alt="Vortex Logo" width="200"/>
</div>

<p align="center">
  <img src="https://img.shields.io/badge/Vortex-V8.1%20Ascended%20Core-8A2BE2?style=for-the-badge" alt="Vortex Lang" width="200">
</p>

<h1 align="center">Vortex Lang v3</h1>

<div align="center">

[![Version](https://img.shields.io/badge/VERSION-v3.0.0-blue?style=for-the-badge&logo=visualstudiocode)](https://marketplace.visualstudio.com/items?itemName=tanmaya2713.vortex-official)
[![Status](https://img.shields.io/badge/STATUS-Public_Released-success?style=for-the-badge)](#)
[![Stars](https://img.shields.io/github/stars/tanmaya2713/VORTEX-LANG?style=for-the-badge)](https://github.com/tanmaya2713/VORTEX-LANG/stargazers)
[![Engine Core](https://img.shields.io/badge/CORE-V8.1_ASCENDED-8A2BE2?style=for-the-badge)](#)
[![License](https://img.shields.io/github/license/tanmaya2713/VORTEX-LANG?style=for-the-badge)](#)

</div>

<p align="center">
  <em>V8.1 Ascended Core — A compiled systems programming language powered by a Go frontend and an LLVM/Clang backend. AI-native, cross-platform.</em>
</p>

---

> **⚠️ IMPORTANT NOTICE: Development Hiatus (2026 - 2028)**
>
> Hey everyone, Tanmaya here. I am taking a strict 2-year hiatus from active development on VORTEX-LANG to focus entirely on my higher secondary studies and competitive entrance exams. The repository will remain public, and the current cross-platform releases are fully stable for you to download, test, and use. Feel free to fork the project, study the source code, and leave issues or PRs—just be aware that I will not be actively reviewing or merging them until 2028. Thank you to everyone who has supported the project so far. See you on the other side! 🚀

---

## 🌍 Engine Compatibility & Status

VORTEX-LANG has been aggressively engineered for cross-platform compatibility. Below is the current deployment status of the nightly builds:

| Operating System | Architecture Target | Status | Deployment Notes |
| :--- | :--- | :--- | :--- |
| **Linux** | `x86_64` & `aarch64` | 🟢 **Working** | Fully native execution. Installed seamlessly via the official `install.sh` pipeline. |
| **macOS** | `arm64` (Apple Silicon) | 🟢 **Working** | Native M-Series support. Requires Homebrew for dependency fallbacks. |
| **Windows** | `x86_64` | 🟢 **Working** | Native execution on Intel/AMD processors. Runs flawlessly via Prism emulation on Windows 11 Copilot+ ARM PCs. |
| **Android** | `arm64` | 🟡 **Beta Testing** | Core binaries compiled. Currently undergoing stability testing for mobile execution environments. |

---

## 📚 Documentation

Explore the deep technical inner workings of the Vortex engine:
* [Architecture Guide](ARCHITECTURE.md) — Dive into the compiler design, lexer mechanics, and cross-platform architecture.

---

## Installation & Verification

### 1. Install the VS Code Extension

Download **[Vortex Language Support](https://marketplace.visualstudio.com/items?itemName=tanmaya2713.vortex-official)** from the VS Code Marketplace for syntax highlighting, snippets, and language assistance.

```bash
code --install-extension tanmaya2713.vortex-official
```

### 2. Install the Compiler

**Linux / macOS / Android (Termux)**

```bash
curl -sSL https://raw.githubusercontent.com/tanmaya2713/VORTEX-LANG/main/dist/install.sh | bash
```

**Windows (PowerShell)**

```powershell
iwr -useb https://raw.githubusercontent.com/tanmaya2713/VORTEX-LANG/main/dist/install.ps1 | iex
```

---

## Verify Installation

Confirm the compiler binary or extension is correctly installed and print the current version based on your specific environment:

### 1. Linux
```bash
source /home/$USER/.bashrc && vortex --version
```

### 2. macOS
```bash
source ~/.bashrc && vortex --version
```

### 3. Windows
```powershell
vortex --version
```

### 4. VS Code Extension
To verify the extension build, check the list of installed extensions. Run the following command and look for `tanmaya2713.vortex-official@0.0.3`. 

You can filter the output by piping it into `grep vortex` (Linux/Mac) or `findstr vortex` (Windows):

**Linux / macOS:**
```bash
code --list-extensions --show-versions | grep vortex
```

**Windows:**
```powershell
code --list-extensions --show-versions | findstr vortex
```

### REPL Playground

Jump into the interactive terminal shell to test Vortex syntax instantly:

```bash
vortex -repl
# Vortex v0.1.0 — type :quit to exit
# vtx> let x = 42;
# vtx> print(x);
# vtx> :quit
```

Works identically on **Windows** (PowerShell / CMD), **Linux**, **macOS**, and **Android (Termux)**.

---

## Usage

```bash
# Build and run locally
vortex run main.vx

# Cross-compile for Android CLI
vortex build main.vx --target android

# Build .so shared library for Android Studio JNI integration
vortex build main.vx --target android --shared
```

---

## Documentation

### General

`fn main()` is the entry point for the program. The file must end with `}`. Anything outside `fn main()` is treated as global or ignored. Use `import` for external modules and `assert` for compile-time invariant checks.

```vortex
// This is ignored or treated as global

fn main() {
    // Write code here
    assert(1 + 1 == 2);
}

// This too is ignored
```

---

### Variables

Variables are declared using `let`. Use `mut` for mutable variables that can be reassigned.

```vortex
fn main() {
    let a = 10;
    let b = "two";
    let c = 15;
    a = a + 1;
    b = "Vortex";
    c = c * 2;
}
```

---

### Data Types

Numbers, strings, and booleans work like other languages. Tensor types are native for AI workloads. `true` and `false` are the boolean values. Structural types are defined with `struct`.

```vortex
fn main() {
    let a = 10;                  // int
    let b = 10 + (15 * 20);      // int expression
    let c = "hello";             // string
    let d = 'ok';                // string (single quotes)
    let e = true;                // bool
    let f = false;               // bool
    let g: tensor<2,3> = [[1, 2, 3], [4, 5, 6]];  // tensor

    struct Point { x: i32; y: i32; }
    let p = Point { x: 10, y: 20 };
}
```

---

### Built-ins & AI Power

Use `print()` to output anything to the console. Tensor operations are first-class citizens with native type syntax and runtime-accelerated math.

```vortex
fn main() {
    print("Hello World");

    let a = 10;
    let b = 20;
    print(a + b);

    // Native tensor matrix multiplication (using overloaded * operator)
    let t: tensor<2,2> = [[1, 2], [3, 4]];
    let result = t * t;
    print(result);
}
```

Native AI constructs — `model`, `layer`, and `train` — enable declarative neural network definitions:

```vortex
model MyNet {
    layer hidden = dense(128, relu);
    layer output = dense(10, sigmoid);
}

fn main() {
    train model=MyNet, data=dataset, epochs=10;
}
```

---

### V2 Omniversal Core — OS & Web I/O

Vortex V2 introduces operating system level file interaction and live API data fetching — directly from your `.vx` scripts.

```vortex
fn main() {
    // Save variables to disk
    let data = "Hello, file system!";
    save data into "output.txt";

    // Load JSON from disk
    load "config.json" into config;
    print(config);

    // Fetch live JSON from a REST API
    request "https://api.example.com/data" into response;
    print(response);
}
```

The `save` keyword writes any variable (including JSON objects) to a file. The `load` keyword reads files and auto-detects JSON vs plain text. The `request` keyword performs read-only HTTP GET requests and returns parsed JSON.

---

### V3 Ascended Core — JSON, Arrays & Extended Math

V3 brings dictionary-style JSON parsing, native Arrays, and a powerful expanded standard library.

```vortex
fn main() {
    // Native Arrays
    let arr = [1, 2, 3, 4, 5];

    // JSON dictionary access
    load "data.json" into obj;
    print(obj["key"]);

    // Extended math builtins
    let x = 2.0;
    print(sin(x));
    print(cos(x));
    print(tan(x));
    print(sqrt(x));
    print(pow(x));

    // Utility builtins
    print(length("hello"));   // → 5
    print(type_of(42));       // → "int"
    print(str(3.14));         // → "3.14"
}
```

Available builtins: `random`, `sqrt`, `round`, `abs`, `pow`, `sin`, `cos`, `tan`, `length`, `type_of`, `str`.

---

### Conditionals

Vortex supports `if`, `else if`, and `else` blocks.

```vortex
fn main() {
    let a = 10;

    if (a < 20) {
        print("a is less than 20");
    } else if (a < 25) {
        print("a is less than 25");
    } else {
        print("a is greater than or equal to 25");
    }
}
```

---

### Loops

Statements inside a `while` block execute as long as the condition evaluates to `true`. Use `break` to exit the loop and `continue` to skip to the next iteration.

```vortex
fn main() {
    let a = 0;

    while (a < 10) {
        a = a + 1;

        if (a == 5) {
            print("skipping five");
            continue;
        }

        if (a == 8) {
            break;
        }

        print(a);
    }

    print("done");
}
```

Vortex also supports `for` loops for iterating over array or tensor elements with an index variable:

```vortex
fn main() {
    for i in [1, 2, 3, 4, 5] {
        print(i);
    }
}
```

> **Note:** The `for` loop range expression must evaluate to an array or tensor type. Scalar literals are not valid ranges.

---

<h2>Roadmap & Future Updates 🚀</h2>

<p>The Vortex Engine is constantly evolving. Here is a sneak peek at what is coming in future releases:</p>

<ul>
  <li><strong>Compiler Enhancements:</strong> We are bringing native tensor and neural net operations directly into the Python interpreter, adding multi-file <code>import</code> statements for massive projects, and building advanced stack traces for pinpoint error debugging.</li>
  <li><strong>Visual GUI Environment:</strong> A dedicated Graphical User Interface (GUI) is in development to let you write, debug, and execute Vortex code visually, extending beyond the command-line CLI.</li>
  <li><strong>Native Package Manager:</strong> A built-in package manager is coming! Soon you'll be able to easily install, publish, and share Vortex modules and ML models with the community.</li>
</ul>

---

<h2>About the Creator &amp; Support</h2>

<p>Vortex is engineered by <strong>Tanmaya Mahapatra (Retired Ame)</strong> under <strong>Drona Labs</strong>.</p>

<ul>
  <li>GitHub: <a href="https://github.com/tanmaya2713">tanmaya2713</a></li>
  <li>Instagram: <a href="https://www.instagram.com/retired_ame/">Retired Ame</a></li>
  <li>Support: <a href="mailto:dronalabs.support@gmail.com">dronalabs.support@gmail.com</a></li>
</ul>

<hr>

<p align="center">
  <em>Built with ❤️ using Go and LLVM</em>
</p>
