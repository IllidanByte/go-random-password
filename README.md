# rgp - Random Password Generator

[English](README.md) | [简体中文](README.zh-CN.md)

`rgp` is a command-line random password generator written in Go. It uses
`crypto/rand` for cryptographically secure randomness and provides both regular
random generation and strong password generation.

## Features

- Combine digits, lowercase letters, uppercase letters, symbols, and safe symbols
- Configure password length and output count
- `gen` command: uniformly samples characters from the selected character sets
- `strong` command: guarantees required character-set coverage and validates
  length, entropy, weak passwords, sequential characters, and repeated characters
- Uses `crypto/rand` throughout for production-friendly randomness
- Ships as a single binary with no runtime dependencies

## Installation

### Build From Source

```bash
# Requires Go >= 1.20
make build
# Binary output: out/rgp
```

### Build With Docker

Use this when you do not want to install Go locally.

```bash
make docker-build
# Linux amd64 binary output: out/rgp
```

### Build All Platforms

```bash
make build-all
# Output: out/<os>_<arch>/rgp[.exe]
# Supported: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
```

### Install With Go

```bash
go install github.com/IllidanByte/go-random-password/cmd/rgp@latest
```

## Usage

```bash
rgp <command> [flags]
```

### `gen` - Regular Random Passwords

`gen` randomly samples characters from the merged character set. It does not
guarantee that every enabled character set appears in the generated password.

```bash
rgp gen [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--length` | `-l` | `20` | Password length |
| `--count` | `-c` | `1` | Number of passwords to generate |
| `--number` | - | `true` | Include digits |
| `--lower` | - | `true` | Include lowercase letters |
| `--upper` | - | `true` | Include uppercase letters |
| `--special` | - | `false` | Include symbols; mutually exclusive with `--special-safe` |
| `--special-safe` | - | `false` | Include safe symbols; mutually exclusive with `--special` |

```bash
# Generate one 20-character password using digits + lowercase + uppercase
rgp gen

# Generate five 16-character passwords
rgp gen --length 16 --count 5

# Include symbols
rgp gen --special

# Include safe symbols
rgp gen --special-safe

# Use digits only
rgp gen --lower false --upper false

# Use lowercase letters only
rgp gen --number false --upper false
```

### `strong` - Strong Passwords

`strong` always enables digits, lowercase letters, and uppercase letters. You can
optionally add one symbol set. Generated passwords must pass all checks below:

1. Length is at least 8 characters
2. Every enabled character set appears at least once
3. Password is not in the built-in weak password list
4. No ASCII sequential run of 3 or more characters, such as `abc` or `123`
5. No repeated run of 3 or more identical characters, such as `aaa` or `111`
6. Entropy is at least 60 bits

```bash
rgp strong [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--length` | `-l` | `20` | Password length; minimum 8 |
| `--count` | `-c` | `1` | Number of passwords to generate |
| `--special` | - | `false` | Add the symbol character set; mutually exclusive with `--special-safe` |
| `--special-safe` | - | `false` | Add the safe symbol character set; mutually exclusive with `--special` |

```bash
# Generate one strong password using digits + lowercase + uppercase
rgp strong

# Generate three strong passwords
rgp strong --count 3

# Add symbols
rgp strong --special

# Add safe symbols and set the length
rgp strong --special-safe --length 16

# Too little entropy returns an error with a suggested minimum length
rgp strong --length 8
```

## Entropy Levels

| Level | Entropy |
|-------|---------|
| Weak | < 40 bits |
| Medium | 40-60 bits |
| Strong | 60-80 bits |
| Very strong | >= 80 bits |

## Character Sets

| Set | Characters |
|-----|------------|
| Digits | `0123456789` |
| Lowercase | `abcdefghijklmnopqrstuvwxyz` |
| Uppercase | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| Symbols | `` `~!@#$%^&*()[{]}-_=+|;:'",<.>/? `` |
| Safe symbols | `-@#%^_+=.,` |

> `gen` only guarantees that characters come from the selected character sets.
> `strong` additionally guarantees character-set coverage and validates password
> strength.

## Library Usage

The reusable core package lives in `password/` and does not depend on the CLI
parser.

```go
package main

import (
	"fmt"
	"log"

	"github.com/IllidanByte/go-random-password/password"
)

func main() {
	pwd, err := password.GenerateStrong(20, password.StrongConfig{
		SpecialSafe: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pwd)
}
```

Public APIs:

```go
func Generate(length int, cfg GenConfig) (string, error)
func GenerateN(length, n int, cfg GenConfig) ([]string, error)
func GenerateStrong(length int, cfg StrongConfig) (string, error)
func GenerateStrongN(length, n int, cfg StrongConfig) ([]string, error)
func Assess(pwd string) StrengthResult
func CalcEntropy(pwd string) float64
```

## Project Layout

```text
.
├── password/         # Reusable core library
│   ├── password.go       # Character sets and generation APIs
│   ├── strength.go       # Strength assessment
│   └── weakpasswords.go  # Built-in weak password list
├── cmd/rgp/          # CLI entry point
│   └── main.go           # gen / strong commands
├── go.mod            # Go module
├── Makefile          # Build scripts
├── Dockerfile        # Docker multi-stage build
└── out/              # Build output
```

## Development

```bash
# Run all tests
go test ./...

# Run static checks
go vet ./...
gofmt -l .

# Show the current injected version
make version
```

## Dependency

- [alecthomas/kong](https://github.com/alecthomas/kong) - command-line parser
