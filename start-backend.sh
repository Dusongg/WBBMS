#!/bin/bash

echo "=================================="
echo "🚀 启动 Go 后端服务"
echo "=================================="
echo ""

cd /Users/dusong/GolandProjects/bookadmin

# 检查 Docker 服务
if ! docker ps | grep -q "bookadmin-mysql"; then
    echo "❌ MySQL 未运行，请先执行: ./start.sh"
    exit 1
fi

if ! docker ps | grep -q "bookadmin-redis"; then
    echo "❌ Redis 未运行，请先执行: ./start.sh"
    exit 1
fi

echo "✅ Docker 服务运行中"
echo ""
echo "🚀 启动 Go 服务..."
echo ""

go run main.go

