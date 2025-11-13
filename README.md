# 图书管理系统

基于 Go 和 Vue 3 构建的简易图书管理系统，使用 gin-vue-admin 框架风格实现。

## 功能特性

- ✅ 图书的增删改查（CRUD）
- ✅ 前端展示、搜索和分页
- ✅ RESTful API 接口
- ✅ MySQL 数据库支持
- ✅ 前后端分离架构

## 技术栈

### 后端
- Go 1.23+
- Gin Web 框架
- GORM ORM
- MySQL 数据库
- Zap 日志库

### 前端
- Vue 3
- Element Plus UI 组件库
- Axios HTTP 客户端
- Vue Router

## 项目结构

```
bookadmin/
├── api/              # API 接口层
│   └── v1/
│       └── book.go   # 图书相关接口
├── model/            # 数据模型
│   ├── book.go       # 图书模型
│   └── common/       # 通用模型
├── router/           # 路由配置
│   ├── book.go       # 图书路由
│   └── router.go     # 路由初始化
├── initialize/       # 初始化模块
│   ├── gorm.go       # 数据库初始化
│   └── logger.go     # 日志初始化
├── global/           # 全局变量
│   └── global.go
├── config.yaml       # 配置文件
├── main.go           # 入口文件
├── go.mod            # Go 模块文件
└── web/              # 前端项目
    ├── src/
    │   ├── views/    # 页面组件
    │   ├── router/   # 路由配置
    │   └── main.js   # 入口文件
    └── package.json  # 前端依赖
```

## 快速开始

### 前置要求

- Go 1.23 或更高版本
- Docker 和 Docker Compose（推荐使用 Docker 启动 MySQL）
- 或 MySQL 5.7+ 或 MySQL 8.0+（本地安装）
- Node.js 16+ 和 npm/yarn

### 1. 数据库准备

#### 方式一：使用 Docker 启动 MySQL（推荐）

使用项目提供的 `docker-compose.yml` 文件启动 MySQL：

```bash
docker-compose up -d
```

这将启动一个 MySQL 8.0 容器，配置如下：
- 容器名称：`bookadmin-mysql`
- 端口：`3306`
- 数据库名：`bookadmin`
- 用户名：`root`
- 密码：`root`
- 字符集：`utf8mb4`

查看容器状态：
```bash
docker-compose ps
```

查看日志：
```bash
docker-compose logs -f mysql
```

停止容器：
```bash
docker-compose down
```

停止并删除数据卷（会删除所有数据）：
```bash
docker-compose down -v
```

#### 方式二：本地 MySQL 安装

创建 MySQL 数据库：

```sql
CREATE DATABASE bookadmin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 配置数据库连接

数据库连接配置在 `initialize/gorm.go` 文件中。如果使用 Docker 启动的 MySQL，默认配置已经匹配，无需修改。

如果需要修改，编辑 `initialize/gorm.go` 文件：

```go
dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
    "root",        // 数据库用户名
    "root",        // 数据库密码
    "127.0.0.1:3306",  // 数据库地址
    "bookadmin",   // 数据库名称
    "charset=utf8mb4&parseTime=True&loc=Local",
)
```

### 3. 安装后端依赖

```bash
go mod tidy
```

### 4. 启动后端服务

```bash
go run main.go
```

后端服务将在 `http://localhost:8888` 启动。

### 5. 安装前端依赖

```bash
cd web
npm install
```

### 6. 启动前端服务

```bash
npm run serve
```

前端服务将在 `http://localhost:8080` 启动。

## API 接口文档

### 图书相关接口

#### 1. 获取图书列表（支持搜索和分页）

```
GET /api/book/getBookList
```

**查询参数：**
- `page`: 页码（默认：1）
- `pageSize`: 每页数量（默认：10）
- `keyword`: 搜索关键字（可选，支持书名、作者、出版社、ISBN）

**响应示例：**
```json
{
  "code": 200,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 10
  },
  "msg": "获取成功"
}
```

#### 2. 创建图书

```
POST /api/book/createBook
```

**请求体：**
```json
{
  "title": "书名",
  "author": "作者",
  "publisher": "出版社",
  "publish_date": "2024-01-01",
  "isbn": "978-7-123456-78-9",
  "price": 59.99,
  "description": "图书描述"
}
```

#### 3. 更新图书

```
PUT /api/book/updateBook
```

**请求体：**
```json
{
  "id": 1,
  "title": "书名",
  "author": "作者",
  ...
}
```

#### 4. 删除图书

```
DELETE /api/book/deleteBook/:id
```

或

```
DELETE /api/book/deleteBook
```

**请求体：**
```json
{
  "id": 1
}
```

#### 5. 获取单个图书

```
GET /api/book/getBook?id=1
```

## 数据库表结构

### books 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键ID |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间（软删除） |
| title | varchar | 书名（必填） |
| author | varchar | 作者（必填） |
| publisher | varchar | 出版社 |
| publish_date | varchar | 出版日期 |
| isbn | varchar | ISBN（唯一索引） |
| price | decimal | 价格 |
| description | text | 描述 |

## 开发说明

### 后端开发

1. 添加新的模型：在 `model/` 目录下创建新的模型文件
2. 添加新的接口：在 `api/v1/` 目录下创建新的接口文件
3. 注册路由：在 `router/` 目录下注册新的路由

### 前端开发

1. 添加新页面：在 `web/src/views/` 目录下创建新的 Vue 组件
2. 配置路由：在 `web/src/router/index.js` 中添加路由配置
3. API 调用：使用 `axios` 调用后端接口

## 注意事项

1. 确保 MySQL 服务已启动
2. 确保数据库已创建
3. 首次运行会自动创建数据表
4. 前端需要配置正确的后端 API 地址
5. 跨域问题已在后端配置，无需额外设置

## 常见问题

### 1. 数据库连接失败

检查：
- MySQL 服务是否启动
- 数据库连接信息是否正确
- 数据库是否已创建

### 2. 前端无法访问后端

检查：
- 后端服务是否启动（端口 8888）
- 前端代理配置是否正确
- 浏览器控制台是否有错误信息

### 3. 依赖安装失败

尝试：
```bash
# Go 依赖
go clean -modcache
go mod tidy

# 前端依赖
cd web
rm -rf node_modules package-lock.json
npm install
```

## 许可证

MIT License

