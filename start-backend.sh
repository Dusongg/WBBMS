#!/bin/bash

set -e

MODE="${1:-all}"

echo "=================================="
echo "启动 Go 后端服务"
echo "=================================="
echo ""

cd /Users/dusong/GolandProjects/bookadmin

if ! docker ps --format '{{.Names}}' | grep -q "^bookadmin-mysql$"; then
    echo "MySQL 未运行，请先执行: ./start.sh"
    exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -q "^bookadmin-redis$"; then
    echo "Redis 未运行，请先执行: ./start.sh"
    exit 1
fi

echo "Docker 依赖已就绪"
echo "启动模式: ${MODE}"
echo ""

go run main.go -mode "${MODE}"

