# 快速启动指南

## 使用 Docker MySQL 启动项目

### 1. 启动 MySQL 容器

```bash
docker-compose up -d
```

等待几秒钟让 MySQL 完全启动，可以查看日志确认：

```bash
docker-compose logs -f mysql
```

看到 `ready for connections` 表示 MySQL 已就绪。

### 2. 安装后端依赖

```bash
go mod tidy
```

### 3. 启动后端服务

```bash
go run main.go
```

后端服务将在 `http://localhost:8888` 启动。

### 4. 安装前端依赖

打开新的终端窗口：

```bash
cd web
npm install
```

### 5. 启动前端服务

```bash
npm run serve
```

前端服务将在 `http://localhost:8080` 启动。

### 6. 访问系统

在浏览器中打开 `http://localhost:8080` 即可使用图书管理系统。

## Docker MySQL 常用命令

### 查看容器状态
```bash
docker-compose ps
```

### 查看日志
```bash
docker-compose logs -f mysql
```

### 停止容器
```bash
docker-compose down
```

### 停止并删除数据（会清空所有数据）
```bash
docker-compose down -v
```

### 重启容器
```bash
docker-compose restart mysql
```

### 进入 MySQL 容器
```bash
docker exec -it bookadmin-mysql bash
```

### 连接 MySQL（在容器内）
```bash
mysql -uroot -proot bookadmin
```

### 直接执行 SQL 命令（推荐）

#### 查看 books 表结构
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "DESCRIBE books;"
```

#### 查看数据总数
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT COUNT(*) as total FROM books;"
```

#### 查看所有图书数据
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT * FROM books;"
```

#### 查看前 10 条数据（格式化输出）
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT * FROM books LIMIT 10\G"
```

#### 查看指定字段的数据
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT id, title, author, publisher, isbn, price, created_at FROM books WHERE deleted_at IS NULL;"
```

#### 使用便捷脚本检查数据
```bash
chmod +x database/check_data.sh
./database/check_data.sh
```

## 数据库配置

Docker MySQL 默认配置：
- **主机**: 127.0.0.1
- **端口**: 3306
- **数据库名**: bookadmin
- **用户名**: root
- **密码**: root

这些配置已经在 `initialize/gorm.go` 中设置好，无需修改。

## 故障排查

### MySQL 容器无法启动

1. 检查端口 3306 是否被占用：
```bash
lsof -i :3306
```

2. 查看容器日志：
```bash
docker-compose logs mysql
```

3. 删除旧容器和数据卷后重新启动：
```bash
docker-compose down -v
docker-compose up -d
```

### 后端无法连接数据库

1. 确认 MySQL 容器正在运行：
```bash
docker-compose ps
```

2. 确认 MySQL 已完全启动（查看日志）：
```bash
docker-compose logs mysql | grep "ready for connections"
```

3. 测试数据库连接：
```bash
docker exec -it bookadmin-mysql mysql -uroot -proot -e "SHOW DATABASES;"
```

### 前端无法访问后端

1. 确认后端服务已启动（端口 8888）
2. 检查浏览器控制台是否有错误
3. 确认 `web/vue.config.js` 中的代理配置正确

