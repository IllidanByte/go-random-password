# rgp — 随机密码生成工具

基于 Go 语言实现的命令行随机密码生成工具，功能参考 FeHelper 随机密码生成器。使用 `crypto/rand` 提供密码学安全的随机性。

## 功能特性

- 支持数字、大小写字母、特殊符号、安全特殊符号五种字符集，可自由组合
- 可指定密码长度与生成数量
- `gen` 子命令：纯随机模式，字符均匀分布
- `strong` 子命令：强密码模式，保证每个字符集各出现至少一次，并通过长度、熵值、弱密码字典、连续/重复字符等多项检查
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
rgp <子命令> [参数]
```

### gen — 普通随机密码

字符从合并字符集中均匀随机抽取，不对结果做强度限制。

```bash
rgp gen [参数]
```

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--length` | `-l` | `20` | 密码长度 |
| `--count` | `-c` | `1` | 生成密码数量 |
| `--number` | — | `true` | 是否包含数字 |
| `--lower` | — | `true` | 是否包含小写字母 |
| `--upper` | — | `true` | 是否包含大写字母 |
| `--special` | — | `false` | 是否包含特殊符号（与 `--special-safe` 互斥）|
| `--special-safe` | — | `false` | 是否包含安全特殊符号（与 `--special` 互斥）|

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

### strong — 强密码模式

默认启用数字 + 小写字母 + 大写字母，可追加特殊字符集。生成的密码保证通过以下所有检查：

1. **长度** ≥ 8 位
2. **字符集覆盖** — 每个启用的字符集各出现至少一个字符
3. **弱密码字典** — 不在内置常见弱密码列表中
4. **无连续字符** — 不含 3 个及以上 ASCII 连续字符（如 `abc`、`123`）
5. **无重复字符** — 不含 3 个及以上相同连续字符（如 `aaa`、`111`）
6. **信息熵** ≥ 60 bits（达到"强"级）

```bash
rgp strong [参数]
```

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--length` | `-l` | `20` | 密码长度（最小 8 位） |
| `--count` | `-c` | `1` | 生成密码数量 |
| `--special` | — | `false` | 追加特殊符号字符集（与 `--special-safe` 互斥）|
| `--special-safe` | — | `false` | 追加安全特殊符号字符集（与 `--special` 互斥）|

```bash
# 生成 1 个强密码（数字 + 小写 + 大写，默认 20 位）
rgp strong

# 生成 3 个强密码
rgp strong --count 3

# 追加特殊符号
rgp strong --special

# 追加安全特殊符号
rgp strong --special-safe --length 16

# 长度不足时工具会报错并给出建议
rgp strong --length 8
# 错误：当前参数信息熵不足（47.6 bits），需 ≥ 60 bits，建议最小长度 11 位
```

### 信息熵评级

| 等级 | 熵值 |
|------|------|
| 弱 | < 40 bits |
| 一般 | 40 – 60 bits |
| **强** | 60 – 80 bits |
| **极强** | ≥ 80 bits |

## 字符集说明

| 字符集 | 内容 |
|--------|------|
| 数字 | `0123456789` |
| 小写字母 | `abcdefghijklmnopqrstuvwxyz` |
| 大写字母 | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| 特殊符号 | `` `~!@#$%^&*()[{]}-_=+|;:'",<.>/? `` |
| 安全特殊符号 | `-@#%^_+=.,` |

> **gen 与 strong 的区别**：`gen` 仅保证字符来自指定字符集，不保证每种字符都出现；`strong` 额外保证每个字符集各出现至少一次，并通过完整的强度检查。

## 项目结构

```
.
├── main.go           # 主程序（CLI 入口、gen/strong 子命令、generatePassword）
├── strength.go       # 强密码检测模块（强度评估、generateStrongPassword）
├── weakpasswords.go  # 内置常见弱密码字典
├── main_test.go      # CLI 与核心函数测试
├── strength_test.go  # 强度模块测试
├── go.mod            # Go 模块依赖
├── Makefile          # 构建脚本
├── Dockerfile        # Docker 多阶段构建
└── out/              # 构建输出目录
```

## 依赖

- [alecthomas/kong](https://github.com/alecthomas/kong) — 命令行参数解析
