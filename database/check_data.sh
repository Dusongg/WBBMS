#!/bin/bash

# 检查 MySQL 容器中的数据
# 使用方法: ./database/check_data.sh

echo "=== 检查 MySQL 容器状态 ==="
docker-compose ps mysql

echo ""
echo "=== 查看 books 表结构 ==="
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "DESCRIBE books;"

echo ""
echo "=== 查看 books 表数据总数 ==="
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT COUNT(*) as total FROM books;"

echo ""
echo "=== 查看 books 表前 10 条数据 ==="
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT * FROM books LIMIT 10\G"

echo ""
echo "=== 查看所有数据（包含软删除） ==="
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT id, title, author, publisher, isbn, price, created_at, updated_at, deleted_at FROM books;"
