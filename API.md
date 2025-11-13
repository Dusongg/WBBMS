# API 接口文档

## 认证相关

### 1. 用户登录
```
POST /api/auth/login
```
**请求体：**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应：**
```json
{
  "code": 200,
  "data": {
    "token": "jwt_token_string",
    "user_id": 1,
    "username": "admin",
    "role": "admin",
    "real_name": "系统管理员"
  },
  "msg": "操作成功"
}
```

### 2. 用户注册
```
POST /api/auth/register
```
**请求体：**
```json
{
  "username": "reader001",
  "password": "password123",
  "email": "reader@example.com",
  "phone": "13800138000",
  "real_name": "张三",
  "id_card": "110101199001011234",
  "address": "北京市朝阳区"
}
```

### 3. 获取用户信息
```
GET /api/auth/userInfo
```
**请求头：**
```
Authorization: Bearer {token}
```

## 图书管理

### 1. 获取图书列表（公开接口）
```
GET /api/book/getBookList?page=1&pageSize=10&keyword=关键词
```

### 2. 获取图书详情（公开接口）
```
GET /api/book/getBook?id=1
```

### 3. 创建图书（需要管理员或图书管理员权限）
```
POST /api/book/createBook
```
**请求头：**
```
Authorization: Bearer {token}
```
**请求体：**
```json
{
  "title": "图书名称",
  "author": "作者",
  "publisher": "出版社",
  "publish_date": "2024-01-01",
  "isbn": "978-7-123456-78-9",
  "price": 59.99,
  "description": "描述",
  "category": "分类",
  "total_stock": 10,
  "available_stock": 10
}
```

### 4. 更新图书
```
PUT /api/book/updateBook
```

### 5. 删除图书
```
DELETE /api/book/deleteBook/:id
```

## 读者管理

### 1. 获取读者列表（需要管理员或图书管理员权限）
```
GET /api/reader/getReaderList?page=1&pageSize=10&keyword=关键词
```

### 2. 获取读者信息
```
GET /api/reader/getReader?id=1
```

### 3. 更新读者状态（审核）
```
PUT /api/reader/updateReaderStatus
```
**请求体：**
```json
{
  "id": 1,
  "status": "active",
  "remark": "审核通过"
}
```
**状态值：** `pending`（待审核）、`active`（正常）、`inactive`（停用）、`rejected`（已拒绝）

### 4. 更新读者信息
```
PUT /api/reader/updateReader
```

## 借还管理

### 1. 借书（需要管理员或图书管理员权限）
```
POST /api/borrow/borrowBook
```
**请求体：**
```json
{
  "reader_id": 1,
  "book_id": 1
}
```

### 2. 还书（需要管理员或图书管理员权限）
```
POST /api/borrow/returnBook
```
**请求体：**
```json
{
  "record_id": 1
}
```

### 3. 续借
```
POST /api/borrow/renewBook
```
**请求体：**
```json
{
  "record_id": 1
}
```

### 4. 获取借阅记录列表（需要管理员或图书管理员权限）
```
GET /api/borrow/getBorrowList?page=1&pageSize=10&keyword=关键词
```

### 5. 获取我的借阅记录
```
GET /api/borrow/getMyBorrowList?page=1&pageSize=10
```

## 统计查询

### 1. 获取统计信息（需要管理员或图书管理员权限）
```
GET /api/statistics/getStatistics
```
**响应：**
```json
{
  "code": 200,
  "data": {
    "total_books": 100,
    "available_books": 80,
    "total_readers": 50,
    "borrowing_count": 20,
    "overdue_count": 5,
    "month_borrow_count": 30,
    "month_return_count": 25
  }
}
```

### 2. 获取借阅统计
```
GET /api/statistics/getBorrowStatistics?start_date=2024-01-01&end_date=2024-12-31
```

### 3. 获取热门图书
```
GET /api/statistics/getPopularBooks
```

## 系统管理

### 1. 获取用户列表（需要管理员权限）
```
GET /api/system/getUserList?page=1&pageSize=10&keyword=关键词
```

### 2. 创建用户（需要管理员权限）
```
POST /api/system/createUser
```
**请求体：**
```json
{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com",
  "phone": "13800138000",
  "role": "reader",
  "real_name": "新用户",
  "status": "active"
}
```
**角色值：** `admin`（系统管理员）、`librarian`（图书管理员）、`reader`（普通读者）

### 3. 更新用户（需要管理员权限）
```
PUT /api/system/updateUser
```

### 4. 删除用户（需要管理员权限）
```
DELETE /api/system/deleteUser
```

## 默认账户

系统初始化时会创建以下默认账户：

- **系统管理员**
  - 用户名：`admin`
  - 密码：`admin123`
  - 角色：系统管理员

- **图书管理员**
  - 用户名：`librarian`
  - 密码：`librarian123`
  - 角色：图书管理员

## 权限说明

- **公开接口**：无需认证，任何人都可以访问
  - 获取图书列表
  - 获取图书详情

- **需要登录**：需要JWT token，登录用户可访问
  - 获取我的借阅记录
  - 续借图书
  - 获取用户信息

- **需要管理员或图书管理员权限**：
  - 图书的增删改
  - 读者管理
  - 借还管理
  - 统计查询

- **需要系统管理员权限**：
  - 用户管理
  - 系统配置

