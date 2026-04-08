BINARY    := rgp
OUT_DIR   := out
MODULE    := github.com/IllidanByte/go-random-password
IMAGE     := rgp-builder

# 当前主机 OS 和架构
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 多架构目标列表
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# ---------- 语义化版本 ----------
# 优先取最近的 semver tag（如 v1.2.3），否则回退到 git describe，再否则用 dev
GIT_TAG    := $(shell git describe --tags --match 'v[0-9]*' --abbrev=0 2>/dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY  := $(shell git status --porcelain 2>/dev/null | grep -q . && echo "-dirty" || echo "")

ifeq ($(GIT_TAG),)
  VERSION := dev-$(GIT_COMMIT)$(GIT_DIRTY)
else
  # tag 存在时：若有未提交改动则追加 -dirty，否则直接用 tag
  VERSION := $(GIT_TAG)$(GIT_DIRTY)
endif

# 仓库有未提交改动时警告（不阻断构建）
ifneq ($(GIT_DIRTY),)
  $(warning 警告：当前工作区存在未提交改动，版本号将标记为 $(VERSION))
endif

# ---------- 构建标志 ----------
# 去除调试符号与 DWARF 信息，缩减二进制体积；同时注入版本信息
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build build-all docker-build clean version

## 默认目标：构建当前主机架构的二进制
all: build

## 打印当前版本号
version:
	@echo $(VERSION)

build:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY) .
	@echo "已生成 $(OUT_DIR)/$(BINARY)（$(GOOS)/$(GOARCH)，版本 $(VERSION)）"

## 构建全部平台二进制，输出至 out/<os>_<arch>/rgp[.exe]
build-all:
	@mkdir -p $(OUT_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS   = $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH = $(word 2,$(subst /, ,$(platform)))) \
		$(eval EXT  = $(if $(filter windows,$(OS)),.exe,)) \
		$(eval DEST = $(OUT_DIR)/$(OS)_$(ARCH)/$(BINARY)$(EXT)) \
		echo "构建 $(platform) -> $(DEST)" && \
		mkdir -p $(OUT_DIR)/$(OS)_$(ARCH) && \
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o $(DEST) . && \
	) true

## 使用 Docker 构建 Linux amd64 二进制（无需本地 Go 环境）
docker-build:
	@mkdir -p $(OUT_DIR)
	docker build --target export -o $(OUT_DIR) \
		--build-arg GOOS=linux --build-arg GOARCH=amd64 \
		--build-arg VERSION=$(VERSION) \
		-f Dockerfile .
	@echo "Docker 构建完成：$(OUT_DIR)/$(BINARY)（版本 $(VERSION)）"

clean:
	rm -rf $(OUT_DIR)
