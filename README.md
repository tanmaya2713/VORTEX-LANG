<p align="center">
  <img src="https://img.shields.io/badge/Vortex-0.1.0--alpha-8A2BE2?style=for-the-badge" alt="Vortex Lang" width="200">
</p>

<h1 align="center">Vortex Lang</h1>

<p align="center">
  <img src="https://github.com/tanmaya2713/VORTEX-LANG/actions/workflows/release.yml/badge.svg" alt="Build Status">
  &nbsp;
  <img src="https://img.shields.io/github/stars/tanmaya2713/VORTEX-LANG?style=flat-square" alt="Stars">
  &nbsp;
  <img src="https://img.shields.io/github/license/tanmaya2713/VORTEX-LANG?style=flat-square" alt="License">
</p>

<p align="center">
  <em>An AI-native, cross-platform systems programming language.</em>
</p>

---

## Installation

**Linux / macOS / Android (Termux)**

```bash
curl -sSL https://raw.githubusercontent.com/tanmaya2713/VORTEX-LANG/main/dist/install.sh | bash
```

**Windows (PowerShell)**

```powershell
iwr -useb https://raw.githubusercontent.com/tanmaya2713/VORTEX-LANG/main/dist/install.ps1 | iex
```

---

## Usage

```bash
# Build and run locally
vortex build main.vtx --run

# Cross-compile for Android CLI
vortex build main.vtx --target android

# Build .so shared library for Android Studio JNI integration
vortex build main.vtx --target android --shared
```

---

## Documentation

### General

`fn main()` is the entry point for the program. The file must end with `}`. Anything outside `fn main()` is treated as global or ignored.

```vortex
// This is ignored or treated as global

fn main() {
    // Write code here
}

// This too is ignored
```

---

### Variables

Variables are declared using `let`.

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

Numbers, strings, and booleans work like other languages. Tensor types are native for AI workloads. `true` and `false` are the boolean values.

```vortex
fn main() {
    let a = 10;                  // int
    let b = 10 + (15 * 20);      // int expression
    let c = "hello";             // string
    let d = 'ok';                // string (single quotes)
    let e = true;                // bool
    let f = false;               // bool
    let g: tensor<2,3> = [[1, 2, 3], [4, 5, 6]];  // tensor
}
```

---

### Built-ins & AI Power

Use `print()` to output anything to the console. Tensor operations are first-class citizens.

```vortex
fn main() {
    print("Hello World");

    let a = 10;
    let b = 20;
    print(a + b);

    // Native tensor matrix multiplication
    let t = tensor([1, 2], [3, 4]);
    let result = t.dot(t);
    print(result);
}
```

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

---

<h2>About the Creator &amp; Support</h2>

<p>Vortex is engineered by <strong>Tanmaya Mahapatra (Retired Ame)</strong> under <strong>Drona Labs</strong>.</p>

<ul>
  <li>GitHub: <a href="https://github.com/tanmaya2713">tanmaya2713</a></li>
  <li>Instagram: <a href="https://instagram.com/retired.ame">Retired Ame</a></li>
  <li>Support: <a href="mailto:dronalabs.support@gmail.com">dronalabs.support@gmail.com</a></li>
</ul>

<hr>

<p align="center">
  <em>Built with ❤️ using Go and LLVM</em>
</p>
