# NeCo Language

NeCo Language is an interpreted programming language built in Go. It features a clean, readable syntax and is statically typed for enhanced safety and reliability. NeCo compiles to byte code, which is run on the NeCo Virtual Machine.

# How to Build

1. Install requirements:
     - Go 1.21.5
     - make
     - MakeGo (optional, for packaging)
2. Use `make build`. The compiled binary will be in `build/debug/`.

# Installation

NeCo Lang can be downloaded as a binary, as a package from the [Releases](https://github.com/DanielNos/neco/releases) page or using Go's package manager.

## Installing Using Go

1. Run: `go install github.com/DanielNos/neco@latest`

## Installing Binary

### Linux

1. Download the correct NeCo Lang binary for your system. 
2. Move the package to /usr/local/bin: `mv [binary name] /usr/local/bin/neco`

### Windows

1. Download the correct NeCo Lang binary for your system.
2. Move it to the folder where it should be installed.
3. Open the **Edit the system environment variables** application.
4. Go to the **Advanced** tab and click on the **Environment Variables...** button.
5. Find the **Path** environment variable and click **Edit**.
6. Click **New** and write the path to your installation folder to the new Field.
7. Click **OK**. You may need to restart your terminal or system to apply this change.

## Installing Package

1. Download the correct NeCo Lang package for your package manager.
2. Install it using the package manager:
    - **APT**: `sudo apt install [package path]`
    - **DNF**: `sudo dnf install [package path]`
    - **PacMan**: `sudo pacman -S [package path]`
  
# Quick Start

1. Create a file `main.neco` with the following content:
```
fun entry() {
  printLine("Hello World!")
}
```
2. Run `neco main.neco`.
3. You should see `Hello World!` in your terminal.

# Usage

`neco [action] [flags] (target)`

## Action and Flags

Each action has its own valid flags.

- `help` Prints help.
- `run` Runs NeCo binary.
- `build` Builds a NeCo Language file to a NeCo binary.
  - `-to`, `--tokens` Prints lexed tokens.
  - `-tr`, `--tree` Draws abstract syntax tree.
  - `-i`, `--instructions` Prints generated instructions.
  - `-d`, `--dont-optimize` Compiler won't optimize byte code.
  - `-s`, `--silent` Doesn't produce info messages when possible.
  - `-n`, `--no-log` Doesn't produce any log messages, even if there are errors.
  - `-l (level)`, `--log-level (level)` Sets logging level. Possible values are 0 to 5 or level names.
  - `-o`, `--out` Sets output file path.
  - `-c`, `--constants` Prints constants stored in binary.
- `analyze` Does syntax and semantic analysis on a NeCo Language source file.
  - `-to`, `--tokens` Prints lexed tokens.
  - `-tr`, `--tree` Draws abstract syntax tree.
  - `-d`, `--dontOptimize` Compiler won't optimize byte code.
