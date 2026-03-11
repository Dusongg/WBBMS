#!/bin/bash

set -e

echo "=================================="
echo "图书管理系统环境启动脚本"
echo "=================================="
echo ""

cd /Users/dusong/GolandProjects/bookadmin

echo "步骤1: 启动 MySQL 和 Redis..."
docker compose up -d mysql redis

echo ""
echo "步骤2: 等待 MySQL 就绪..."
until docker exec bookadmin-mysql mysqladmin ping -h 127.0.0.1 -proot --silent >/dev/null 2>&1; do
    sleep 2
done

echo "MySQL 已就绪"
echo ""

echo "步骤3: 执行应用迁移和初始化..."
go run main.go -mode migrate
echo ""

echo "步骤4: 验证 Redis 连接..."
docker exec bookadmin-redis redis-cli ping
echo ""

echo "当前 Docker 服务状态:"
docker compose ps
echo ""

echo "=================================="
echo "环境准备完成"
echo "=================================="
echo ""
echo "启动本地一体模式:"
echo "  ./start-backend.sh"
echo ""
echo "或分别启动:"
echo "  ./start-backend.sh api"
echo "  ./start-backend.sh worker"
echo "  ./start-backend.sh scheduler"
echo ""
echo "如需完整容器化部署:"
echo "  docker compose --profile app run --rm migrate"
echo "  docker compose --profile app up -d api worker scheduler nginx"
echo ""
