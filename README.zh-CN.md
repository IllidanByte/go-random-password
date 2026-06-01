# rgp - 随机密码生成工具

[English](README.md) | [简体中文](README.zh-CN.md)

`rgp` 是一个基于 Go 语言实现的命令行随机密码生成工具。它使用
`crypto/rand` 提供密码学安全的随机性，并同时支持普通随机密码生成与强密码生成。

## 功能特性

- 支持数字、大小写字母、特殊符号、安全特殊符号五种字符集，可自由组合
- 可指定密码长度与生成数量
- `gen` 子命令：纯随机模式，字符均匀分布
- `strong` 子命令：强密码模式，保证每个字符集各出现至少一次，并通过长度、熵值、弱密码字典、连续字符、重复字符等多项检查
- 使用 `crypto/rand` 保证随机性，适合生产环境密码生成
- 单一二进制，无运行时依赖

## 安装

### 从源码构建

```bash
# 需要 Go >= 1.20
make build
# 二进制输出至 out/rgp
```

### 通过 Docker 构建

无需本地 Go 环境时可使用 Docker 构建。

```bash
make docker-build
# 输出 Linux amd64 二进制至 out/rgp
```

### 构建全平台二进制

```bash
make build-all
# 输出至 out/<os>_<arch>/rgp[.exe]
# 支持：linux/amd64、linux/arm64、darwin/amd64、darwin/arm64、windows/amd64
```

### 通过 Go 安装

```bash
go install github.com/IllidanByte/go-random-password/cmd/rgp@latest
```

## 使用方法

```bash
rgp <子命令> [参数]
```

### `gen` - 普通随机密码

`gen` 从合并字符集中均匀随机抽取字符，不保证每个启用字符集都出现在生成结果中。

```bash
rgp gen [参数]
```

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--length` | `-l` | `20` | 密码长度 |
| `--count` | `-c` | `1` | 生成密码数量 |
| `--number` | - | `true` | 是否包含数字 |
| `--lower` | - | `true` | 是否包含小写字母 |
| `--upper` | - | `true` | 是否包含大写字母 |
| `--special` | - | `false` | 是否包含特殊符号，与 `--special-safe` 互斥 |
| `--special-safe` | - | `false` | 是否包含安全特殊符号，与 `--special` 互斥 |

```bash
# 生成 1 个 20 位密码（默认：数字 + 小写 + 大写）
rgp gen

# 生成 5 个 16 位密码
rgp gen --length 16 --count 5

# 启用特殊符号
rgp gen --special

# 启用安全特殊符号
rgp gen --special-safe

# 仅使用数字
rgp gen --lower false --upper false

# 仅使用小写字母
rgp gen --number false --upper false
```

### `strong` - 强密码模式

`strong` 默认启用数字、小写字母、大写字母三类基础字符集，可追加一种特殊字符集。生成的密码必须通过以下检查：

1. 长度至少 8 位
2. 每个启用的字符集各出现至少一个字符
3. 不在内置常见弱密码列表中
4. 不含 3 个及以上 ASCII 连续字符，例如 `abc`、`123`
5. 不含 3 个及以上相同连续字符，例如 `aaa`、`111`
6. 信息熵至少 60 bits

```bash
rgp strong [参数]
```

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--length` | `-l` | `20` | 密码长度，最小 8 位 |
| `--count` | `-c` | `1` | 生成密码数量 |
| `--special` | - | `false` | 追加特殊符号字符集，与 `--special-safe` 互斥 |
| `--special-safe` | - | `false` | 追加安全特殊符号字符集，与 `--special` 互斥 |

```bash
# 生成 1 个强密码（数字 + 小写 + 大写，默认 20 位）
rgp strong

# 生成 3 个强密码
rgp strong --count 3

# 追加特殊符号
rgp strong --special

# 追加安全特殊符号并指定长度
rgp strong --special-safe --length 16

# 信息熵不足时工具会报错并给出建议长度
rgp strong --length 8
```

## 信息熵评级

| 等级 | 熵值 |
|------|------|
| 弱 | < 40 bits |
| 一般 | 40-60 bits |
| 强 | 60-80 bits |
| 极强 | >= 80 bits |

## 字符集说明

| 字符集 | 内容 |
|--------|------|
| 数字 | `0123456789` |
| 小写字母 | `abcdefghijklmnopqrstuvwxyz` |
| 大写字母 | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| 特殊符号 | `` `~!@#$%^&*()[{]}-_=+|;:'",<.>/? `` |
| 安全特殊符号 | `-@#%^_+=.,` |

> `gen` 仅保证字符来自指定字符集；`strong` 额外保证每个字符集各出现至少一次，并通过完整的强度检查。

## 库用法

可复用的核心包位于 `password/`，不依赖 CLI 参数解析库。

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

公开 API：

```go
func Generate(length int, cfg GenConfig) (string, error)
func GenerateN(length, n int, cfg GenConfig) ([]string, error)
func GenerateStrong(length int, cfg StrongConfig) (string, error)
func GenerateStrongN(length, n int, cfg StrongConfig) ([]string, error)
func Assess(pwd string) StrengthResult
func CalcEntropy(pwd string) float64
```

## 项目结构

```text
.
├── password/         # 核心库，可被其他 Go 项目 import
│   ├── password.go       # 字符集常量、Generate / GenerateStrong 等
│   ├── strength.go       # 强密码评估
│   └── weakpasswords.go  # 内置常见弱密码字典
├── cmd/rgp/          # CLI 入口
│   └── main.go           # gen / strong 子命令
├── go.mod            # Go 模块依赖
├── Makefile          # 构建脚本
├── Dockerfile        # Docker 多阶段构建
└── out/              # 构建输出目录
```

## 开发

```bash
# 运行所有测试
go test ./...

# 代码检查
go vet ./...
gofmt -l .

# 查看当前版本号
make version
```

## 依赖

- [alecthomas/kong](https://github.com/alecthomas/kong) - 命令行参数解析
