# rgp — 随机密码生成工具

基于 Go 语言实现的命令行随机密码生成工具，功能参考 FeHelper 随机密码生成器。使用 `crypto/rand` 提供密码学安全的随机性。

## 功能特性

- 支持数字、大小写字母、特殊符号、安全特殊符号五种字符集，可自由组合
- 可指定密码长度与生成数量
- 使用 `crypto/rand` 保证随机性，适合生产环境密码生成
- 单一二进制，无运行时依赖

## 安装

### 从源码构建

```bash
# 需要 Go >= 1.20
make build
# 二进制输出至 out/rgp
```

### 通过 Docker 构建（无需本地 Go 环境）

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

## 使用方法

```bash
rgp [参数]
```

### 参数说明

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--length` | `-l` | `20` | 密码长度 |
| `--count` | `-c` | `1` | 生成密码数量 |
| `--number` | — | `true` | 是否包含数字 |
| `--lower` | — | `true` | 是否包含小写字母 |
| `--upper` | — | `true` | 是否包含大写字母 |
| `--special` | — | `false` | 是否包含特殊符号（与 --special-safe 互斥）|
| `--special-safe` | — | `false` | 是否包含安全特殊符号（与 --special 互斥）|

### 示例

```bash
# 生成 1 个 20 位密码（默认：数字 + 小写 + 大写）
rgp

# 生成 1 个 50 位密码
rgp --length 50

# 生成 5 个 16 位密码
rgp --length 16 --count 5

# 启用特殊符号
rgp --special true

# 启用安全特殊符号（不含引号、括号等易混淆字符）
# 等价写法：--special-safe true 或裸用 --special-safe
rgp --special-safe

# 仅使用数字
rgp --lower false --upper false

# 仅使用小写字母
rgp --number false --upper false

# 数字 + 小写 + 特殊符号（不含大写）
rgp --upper false --special true

# 仅字母（大小写）
rgp --number false
```

## 字符集说明

启用某个字符集意味着该字符集中的字符**有可能**出现在生成的密码中（按均匀随机分布抽取），并不保证每种字符都至少出现一次。

| 字符集 | 内容 |
|--------|------|
| 数字 | `0123456789` |
| 小写字母 | `abcdefghijklmnopqrstuvwxyz` |
| 大写字母 | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| 特殊符号 | `` `~!@#$%^&*()[{]}-_=+|;:'",<.>/? `` |
| 安全特殊符号 | `-@#%^_+=.,` |

## 项目结构

```
.
├── main.go       # 主程序
├── go.mod        # Go 模块依赖
├── Makefile      # 构建脚本
├── Dockerfile    # Docker 多阶段构建
└── out/          # 构建输出目录
```

## 依赖

- [alecthomas/kong](https://github.com/alecthomas/kong) — 命令行参数解析
