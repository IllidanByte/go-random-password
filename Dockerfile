# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/rgp ./cmd/rgp

# 仅用于导出二进制，不产生最终镜像
FROM scratch AS export
COPY --from=builder /out/rgp /rgp
