#!/bin/bash
# Docker 环境数据库迁移脚本

echo "==================================="
echo "Docker MySQL 数据库迁移助手"
echo "==================================="
echo ""

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ 错误: Docker 未运行，请先启动 Docker Desktop"
    exit 1
fi

echo "📋 当前运行的 MySQL 和 Redis 容器:"
echo ""
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Ports}}" | grep -E "mysql|redis|NAME"
echo ""

# 提示用户输入容器名
read -p "请输入 MySQL 容器名称（例如: mysql）: " MYSQL_CONTAINER
read -sp "请输入 MySQL root 密码: " MYSQL_PASSWORD
echo ""
read -p "请输入数据库名称（默认: bookadmin）: " DB_NAME
DB_NAME=${DB_NAME:-bookadmin}

echo ""
echo "开始执行数据库迁移..."
echo ""

# SQL 文件路径
SQL_FILE="/Users/dusong/GolandProjects/bookadmin/database/migration_likes_favorites.sql"

# 方法1: 直接通过管道执行
echo "📦 方法1: 使用 docker exec 直接执行..."
if docker exec -i "$MYSQL_CONTAINER" mysql -u root -p"$MYSQL_PASSWORD" "$DB_NAME" < "$SQL_FILE" 2>/dev/null; then
    echo "✅ 数据库迁移成功！"
    echo ""
    echo "验证迁移结果:"
    docker exec -i "$MYSQL_CONTAINER" mysql -u root -p"$MYSQL_PASSWORD" "$DB_NAME" -e "SHOW TABLES LIKE 'book_%';"
    exit 0
else
    echo "⚠️  方法1失败，尝试方法2..."
    echo ""
    
    # 方法2: 复制文件到容器内再执行
    echo "📦 方法2: 复制SQL文件到容器内..."
    docker cp "$SQL_FILE" "$MYSQL_CONTAINER":/tmp/migration.sql
    
    echo "📦 执行SQL脚本..."
    docker exec -i "$MYSQL_CONTAINER" mysql -u root -p"$MYSQL_PASSWORD" "$DB_NAME" -e "source /tmp/migration.sql"
    
    if [ $? -eq 0 ]; then
        echo "✅ 数据库迁移成功！"
        echo ""
        echo "验证迁移结果:"
        docker exec -i "$MYSQL_CONTAINER" mysql -u root -p"$MYSQL_PASSWORD" "$DB_NAME" -e "SHOW TABLES LIKE 'book_%';"
        
        # 清理临时文件
        docker exec -i "$MYSQL_CONTAINER" rm /tmp/migration.sql
        exit 0
    else
        echo "❌ 迁移失败，请检查以下内容："
        echo "  1. MySQL 容器名是否正确"
        echo "  2. 密码是否正确"
        echo "  3. 数据库 $DB_NAME 是否存在"
        exit 1
    fi
fi

