#!/bin/bash

echo "=================================="
echo "📚 图书管理系统启动脚本"
echo "=================================="
echo ""

# 1. 启动 Docker 服务
echo "🐳 步骤1: 启动 MySQL 和 Redis..."
cd /Users/dusong/GolandProjects/bookadmin
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "❌ Docker 启动失败，请检查 Docker Desktop 是否运行"
    exit 1
fi

echo "✅ Docker 服务启动成功"
echo ""

# 2. 等待 MySQL 完全启动
echo "⏳ 等待 MySQL 启动（10秒）..."
sleep 10

# 3. 执行数据库迁移
echo "📦 步骤2: 执行数据库迁移..."
docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin < /Users/dusong/GolandProjects/bookadmin/database/migration_likes_favorites.sql 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ 数据库迁移成功"
else
    echo "⚠️  数据库迁移可能已执行过，跳过..."
fi
echo ""

# 4. 验证表结构
echo "🔍 验证数据库表..."
docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin -e "SHOW TABLES LIKE 'book_%';"
echo ""

# 5. 测试 Redis 连接
echo "🔍 测试 Redis 连接..."
docker exec -i bookadmin-redis redis-cli ping
echo ""

# 6. 显示服务状态
echo "📊 Docker 服务状态:"
docker-compose ps
echo ""

echo "=================================="
echo "✅ 环境准备完成！"
echo "=================================="
echo ""
echo "下一步: 启动 Go 服务"
echo "  cd /Users/dusong/GolandProjects/bookadmin"
echo "  go run main.go"
echo ""
echo "或者在新终端窗口直接运行:"
echo "  ./start-backend.sh"
echo ""
