## Lesson 01

# Hello, Go

Learning goal: get a Go program running on your machine, understand `package main` and `func main()`, and use `fmt.Println` / `fmt.Printf` to print formatted output.

---

## What we'll cover

- Installing Go and verifying with `go version`.
- Running your first Go file with `go run`.
- Initialising a project with `go mod init` and what `go.mod` is.
- The role of `package main` and `func main()`.
- Printing with `fmt.Println` and `fmt.Printf`.

---

## Concept 1: From Go installed to running code

### Motivation

Before we write Go programs, we need to be able to run them. Unlike scripting languages, Go is compiled — but we don't have to deal with that explicitly for now. Go's tooling handles compilation behind the scenes.

For small programs you can think of `go run file.go` as "run this Go file" the same way you'd think `python file.py` runs a Python script. This first concept is about getting from "Go is installed" to "I just ran my first Go program."

---

### The basics

Save this as `hello.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

Then run:

```bash
go run hello.go
```

You should see:

```
Hello, Go!
```

That's it. You just compiled and ran a Go program.

---

### A worked example

Let's start the running theme of this course: a personal expense tracker. By the end of Phase 1, you'll have built a small CLI for it. For now, just print a hello-world version:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello! You spent €23.50 on coffee today.")
}
```

Run it:

```bash
go run hello.go
```

You'll see:

```
Hello! You spent €23.50 on coffee today.
```

That `€23.50` is hardcoded as a string for now. Lesson 02 introduces variables and types so the `23.50` can be a number you compute with.

---

### Common mistake

A classic mistake: forgetting `package main`.

```go
import "fmt"

func main() {
	fmt.Println("Hello!")
}
```

Run it and Go complains:

```
hello.go:1:1: expected 'package', found 'import'
```

Every Go file starts with a `package` declaration. For executable programs, that line is exactly `package main`.

---

### Recap

- `go run file.go` compiles and runs a Go file in one shot.
- Every Go file starts with `package <something>`.
- An executable program lives in `package main`.

---

## Concept 2: Modules and the project structure

### Motivation

`go run hello.go` works for a single file. But real Go programs are organised into modules — a directory with a `go.mod` file at the top. The module is what `go build`, `go test`, and `go run ./...` understand as "this project."

For Phase 1, every lesson lives inside the course's single module. You'll create modules of your own when you start projects from scratch.

---

### The basics

Create a new directory and initialise a module:

```bash
mkdir hello-project
cd hello-project
go mod init example.com/hello
```

You'll get a `go.mod` file:

```
module example.com/hello

go 1.23
```

The `module` line is the *import path*. The `go` line is the minimum Go version. That's all you need to start.

---

### A worked example

Add a `main.go` next to the `go.mod`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello from a real module!")
}
```

Now you can run:

```bash
go run .
```

The `.` means "the current package" — Go looks at all `.go` files in this directory, compiles them, and runs `main()`. That's the standard way to run programs from a module.

You can also still use `go run main.go` if you prefer.

---

### Common mistake

Confusing the module path with the directory name. The module path (`example.com/hello`) is just an identifier — it tells Go's tooling how to refer to your module if it were imported by someone else. It does not have to match your directory name.

The convention is to use a path you control on the internet — `github.com/yourusername/yourproject` is the most common form. A bare `my-project` works for personal use, but published modules need a real path.

---

### Recap

- A *module* is a directory with a `go.mod` file at the top.
- `go mod init <path>` creates the file.
- `go run .` runs all `.go` files in the current package.
- The module path is a logical identifier, not a directory name.

---

## Concept 3: The main package and main function

### Motivation

Every executable Go program has the same starting point: `func main()` inside `package main`. That's how Go knows which function to call when you run the binary.

This sounds simple, but the *combination* — special package name, special function name — is unique to Go and worth understanding.

---

### The basics

```go
package main

func main() {
	// This runs when the program starts.
}
```

Two rules for executable programs:

1. The package must be named `main`.
2. There must be exactly one function called `main` with no parameters and no return values.

If either is missing, `go run` complains.

---

### A worked example

Let's add some variables (technically lesson 02 territory, but useful here as a preview):

```go
package main

import "fmt"

func main() {
	date := "2026-05-07"
	amount := 4.50
	category := "coffee"
	fmt.Println(date, amount, category)
}
```

Output:

```
2026-05-07 4.5 coffee
```

`Println` separates its arguments with spaces and adds a newline at the end. Notice that `4.50` printed as `4.5` — Go drops trailing zeros for floats by default. We'll fix that with `Printf` in the next concept.

---

### Common mistake

Capitalisation matters. Go is case-sensitive everywhere:

```go
package main

func Main() {  // wrong: capital M
	fmt.Println("Hello!")
}
```

Run it and Go complains:

```
runtime.main_main·f: function main is undeclared in the main package
```

The fix is one keystroke: `Main` → `main`.

---

### Recap

- Executable programs need `package main` and `func main()`.
- `func main()` takes no arguments and returns nothing.
- Go is case-sensitive: `main` is not the same as `Main`.

---

## Concept 4: fmt and basic printing

### Motivation

You'll print things constantly while learning a language: to inspect values, to give feedback, to produce output. The `fmt` package is Go's primary tool for that.

`Println` is the simple "print this with a newline" function. `Printf` is the formatted version — it lets you control how numbers, strings, and other values look in the output. For lesson 01 we focus on these two, plus `Sprintf` (the same as `Printf` but returns a string instead of printing it).

---

### The basics

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")              // prints: Hello, world!
	fmt.Println("two", "things")              // prints: two things
	fmt.Printf("amount = %.2f\n", 4.5)        // prints: amount = 4.50
	fmt.Printf("%s ate %d fish\n", "Aki", 3)  // prints: Aki ate 3 fish
}
```

Three things to notice:

- `Println` joins its arguments with spaces and adds a newline.
- `Printf` uses *format verbs* (the `%` placeholders) but does NOT add a newline — you have to write `\n` yourself.
- `%s` is for strings, `%d` is for integers, `%f` is for floats. `%.2f` is "float with 2 decimal places."

---

### A worked example

The expense-tracker theme — print three expenses with two-decimal amounts:

```go
package main

import "fmt"

func main() {
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 4.50, "coffee")
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 23.50, "groceries")
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-06", 1234.50, "rent")
}
```

Output:

```
2026-05-07  €4.50  coffee
2026-05-07  €23.50  groceries
2026-05-06  €1234.50  rent
```

Every amount is printed with exactly two decimal places, including `4.50` (which would have lost the trailing zero with `Println`). The `%.2f` verb gives consistent formatting.

---

### Common mistake

Forgetting the `\n` at the end of a `Printf`:

```go
fmt.Printf("Hello!")
fmt.Printf("World!")
```

Output:

```
Hello!World!
```

No newlines, no separation. `Println` adds the newline; `Printf` does not. If you want the line to end, include `\n` in the format string.

---

### Recap

- `fmt.Println` for simple "print + newline" output.
- `fmt.Printf` for formatted output. You write the newline yourself.
- `fmt.Sprintf` returns a formatted string instead of printing it.
- Format verbs: `%s` strings, `%d` ints, `%f` floats. `%.2f` for two decimals.

---

## Practice

### Warm-up

Two small functions in `exercises/warmup.go`:

- `WarmupHello()` returns `"Hello, Go!"`.
- `WarmupGreet(name)` returns `"Hello, <name>!"` for any name.

Both are one line of code using string concatenation. Make the failing tests pass.

```bash
cd lessons/01-hello/exercises
go test -run Warmup -v
```

---

### Main

`FormatExpense(date, amount, category)` in `exercises/main.go` returns a column-formatted string like `"2026-05-07  €4.50  coffee"`. Use `fmt.Sprintf` with `%s` and `%.2f` verbs.

The starter doesn't import `fmt` — you'll add `import "fmt"` at the top of `main.go`.

```bash
cd lessons/01-hello/exercises
go test -v
```

Note:
For live: walk through `go run hello.go` on the projector. Hit the common mistakes deliberately so students see the error messages — `package main` missing, capital `Main`, missing `\n`. Skip the longer worked example in concept 2 if running short on time; the exercises cover that ground.

---

## What we learned

- `go run file.go` compiles and runs a Go program in one step.
- A *module* is a directory with `go.mod`; `go mod init` creates it.
- Executable programs are `package main` with `func main()`.
- `fmt.Println` for simple printing; `fmt.Printf` for formatted output; `fmt.Sprintf` returns the formatted string.
- Go is case-sensitive — `main` not `Main`.

---

## Up next

Lesson 02 — Variables, types, operators.
