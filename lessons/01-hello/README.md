# Lesson 01: Hello, Go

## Learning goals

- Verify your Go installation and run a Go program from the command line.
- Initialise a new Go module with `go mod init` and understand what `go.mod` contains.
- Recognise the role of `package main` and `func main()` in executable Go programs.
- Print formatted output with `fmt.Println` and `fmt.Printf`, including float formatting with `%.2f`.

## Prerequisites

This is the first lesson — no prior lessons assumed. You should have:

- Go 1.23 or newer installed (`go version` should print something like `go version go1.23.0 darwin/arm64`).
- A terminal you're comfortable typing in.
- A text editor with Go support — VS Code, GoLand, vim, emacs all work.

If you don't have Go installed yet, the official guide is at <https://go.dev/doc/install>.

## Concepts

### From Go installed to running code

Go is a compiled language, but for everyday work you don't need to think about that explicitly. The `go run` command compiles your file in the background and runs the result. Save a file as `hello.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

Then run it:

```bash
go run hello.go
```

You'll see `Hello, Go!` printed. That's the smallest complete Go program — three required pieces: the package declaration on line 1, the `import "fmt"` to bring in the printing package, and `func main()` as the program's entry point.

The course threads through a small running example — a personal expense tracker. By the end of Phase 1, you'll have built a working CLI for it. For now, just print a hello-world version with the theme:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello! You spent €23.50 on coffee today.")
}
```

That `€23.50` is hardcoded as a string for now. Lesson 02 introduces variables and types so the `23.50` can be a real number you compute with.

**Common mistake.** Forgetting `package main`:

```go
import "fmt"

func main() {
	fmt.Println("Hello!")
}
```

Go complains:

```
hello.go:1:1: expected 'package', found 'import'
```

Every Go file starts with a `package` declaration. For executable programs, that line is exactly `package main`.

### Modules and the project structure

A single `.go` file works for tiny demos, but real Go projects are organised as *modules*. A module is a directory with a `go.mod` file at the top. You create one with:

```bash
go mod init example.com/hello
```

The `go.mod` file looks like this:

```
module example.com/hello

go 1.23
```

`example.com/hello` is the module path — a logical identifier that other modules would use to import yours. `go 1.23` declares the minimum Go version. Inside a module, you can run `go run .` (the dot means "the current package") and Go figures out which files to compile. This is the standard way to run programs once a project has more than one file.

**Common mistake.** Confusing the module path with the directory name. The module path (`example.com/hello`) is just an identifier — it tells Go's tooling how to refer to your module if it were imported by someone else. It does not have to match your directory name. The convention is to use a path you control on the internet — `github.com/yourusername/yourproject` is the most common form.

### The main package and main function

Every executable Go program has the same entry point: `func main()` inside `package main`. The package name `main` is what tells Go this is a runnable program (rather than a library). The `main` function takes no parameters and returns nothing — its only job is to run when the program starts.

If you write `package mainz` or `func Main()` (capital M), Go won't run the program. It'll either complain at compile time or report that no entry point exists. Case matters.

**Common mistake.** Capitalisation matters. Go is case-sensitive everywhere:

```go
package main

import "fmt"

func Main() {  // wrong: capital M
	fmt.Println("Hello!")
}
```

Go complains:

```
runtime.main_main·f: function main is undeclared in the main package
```

The fix is one keystroke: `Main` → `main`.

### Formatting output with fmt

The `fmt` package handles printing. Two functions cover most of what you need in lesson 01:

- `fmt.Println(args...)` — prints its arguments separated by spaces and adds a newline. Best for quick output.
- `fmt.Printf(format, args...)` — uses *format verbs* like `%s`, `%d`, `%f` to control how each argument is printed. You write the newline yourself with `\n`.

The format verbs you'll use most:

- `%s` — a string
- `%d` — a decimal integer
- `%f` — a floating-point number (default formatting, may have many decimals)
- `%.2f` — a float with exactly two decimal places (great for currency)

Example:

```go
fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 4.50, "coffee")
```

prints:

```
2026-05-07  €4.50  coffee
```

`fmt.Sprintf` is the same as `Printf` but returns the formatted string instead of printing it. You'll use `Sprintf` in this lesson's main exercise.

**Common mistake.** Forgetting the `\n` at the end of a `Printf`:

```go
fmt.Printf("Hello!")
fmt.Printf("World!")
```

Output:

```
Hello!World!
```

No newlines, no separation. `Println` adds the newline; `Printf` does not. If you want the line to end, include `\n` in the format string.

## Exercise: warm-up

Open `exercises/warmup.go`. There are two small functions to implement:

- `WarmupHello() string` — return the literal string `"Hello, Go!"`.
- `WarmupGreet(name string) string` — return `"Hello, <name>!"` where `<name>` is the input.

Both are one-line functions using string concatenation (`+`). Make the failing tests in `warmup_test.go` pass.

## Exercise: main

Open `exercises/main.go`. The `Greet` function is already provided as a worked example; your job is `FormatExpense`.

`FormatExpense(date string, amount float64, category string) string` returns a column-formatted line like `"2026-05-07  €4.50  coffee"`. Use `fmt.Sprintf` with the `%s` and `%.2f` format verbs to build the string.

You'll need to add `import "fmt"` at the top of the file — the starter doesn't import it because the panic stub doesn't need it.

The tests cover several cases including amounts with zero decimals (must still print as `.00`), big numbers (no comma separators), and small numbers.

## How to run

```bash
cd lessons/01-hello/exercises
go test -run Warmup -v   # warm-up only
go test -v                # everything
```

After you finish (or while iterating), run `gofmt -w .` from the lesson folder to keep your code formatted in the canonical Go style. Make this a habit — every Go project does it the same way, and the tooling enforces it.

Once both exercises pass, take a look at `solutions/` to compare your code with the reference implementation. The solutions might use slightly different idioms — that's fine.

To experiment with `go run` on your own, create a `hello.go` file outside this folder (e.g., in `~/scratch/`) and try the examples from the slides. The lesson folder itself uses test-driven exercises, not runnable `main()` programs.

## Going further

### Read

- [A Tour of Go — Welcome](https://go.dev/tour/welcome/1) — the official interactive intro. Walks through similar material with playgrounds you can edit live.
- [The fmt package documentation](https://pkg.go.dev/fmt) — the canonical reference for format verbs. Skim the "Printing" section; you'll use it a lot.
- [Effective Go — Names](https://go.dev/doc/effective_go#names) — short read about Go's naming conventions. Worth reading early to internalise the style.

### Try

- **Currency localisation.** Modify your `FormatExpense` to take a currency symbol parameter (e.g. `"€"`, `"$"`, `"£"`) and use it in the output. No reference solution provided — pick the function signature that feels most readable to you.
- **Right-alignment.** Format verbs support width specifiers: `%8.2f` pads the float to width 8, right-aligned. Try printing a small list of expenses where every amount is right-aligned to width 8. The output should look like a clean column.
