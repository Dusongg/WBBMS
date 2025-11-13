# 🚀 图书管理系统启动指南

## 一键启动（推荐）

### 终端1: 启动后端服务

```bash
cd /Users/dusong/GolandProjects/bookadmin
./start-backend.sh
```

**成功标志**：
```
[INFO] Redis连接成功
[INFO] 成功创建Redis Stream消费者组 sync-group
[INFO] 同步Worker [worker-1] 启动
[INFO] 同步Worker [worker-2] 启动
[INFO] 同步Worker [worker-3] 启动
[INFO] 同步Worker [worker-4] 启动
[INFO] 同步Worker [worker-5] 启动
[INFO] Worker池启动完成，共 5 个Worker
[INFO] 服务器启动在端口: 8888
```

---

### 终端2: 启动前端服务

```bash
cd /Users/dusong/GolandProjects/bookadmin/web
npm run serve
```

**成功标志**：
```
  App running at:
  - Local:   http://localhost:8080/
  - Network: http://192.168.x.x:8080/
```

---

## 📱 访问系统

打开浏览器访问：**http://localhost:8080**

**默认账号**：
- 用户名: `admin`
- 密码: `admin123`

---

## 🎯 测试点赞/收藏功能

### 步骤1: 登录系统
1. 打开 http://localhost:8080
2. 输入用户名 `admin`，密码 `admin123`
3. 点击登录

### 步骤2: 进入图书列表
1. 点击顶部导航栏的"图书管理"或"图书列表"
2. 等待图书列表加载完成

### 步骤3: 测试点赞功能
1. **鼠标悬停**在任意图书卡片上
2. **查看底部操作栏**：会淡入显示
3. **点击❤️按钮**：
   - 看到弹跳动画
   - 心形从🤍变成❤️
   - 计数+1
   - 提示"点赞成功 ❤️"
4. **再次点击❤️**：
   - 心形从❤️变回🤍
   - 计数-1
   - 提示"已取消点赞"

### 步骤4: 测试收藏功能
1. **点击⭐按钮**：
   - 看到旋转动画
   - 星星从☆变成⭐
   - 计数+1
   - 提示"收藏成功 ⭐"
2. **再次点击⭐**：
   - 星星从⭐变回☆
   - 计数-1
   - 提示"已取消收藏"

---

## 🔍 验证数据持久化

### 查看Redis缓存

```bash
# 查看所有key
docker exec -i bookadmin-redis redis-cli keys '*'

# 查看用户点赞状态
docker exec -i bookadmin-redis redis-cli SMEMBERS "user:likes:1"

# 查看图书统计
docker exec -i bookadmin-redis redis-cli HGETALL "book:stats:1"

# 查看点赞周榜
docker exec -i bookadmin-redis redis-cli ZREVRANGE "rank:likes:week:2025-W45" 0 10 WITHSCORES
```

### 查看MySQL数据

```bash
# 查看点赞记录
docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT * FROM book_likes LIMIT 10;"

# 查看收藏记录
docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT * FROM book_favorites LIMIT 10;"

# 查看图书统计
docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin -e "SELECT id, title, like_count, favorite_count FROM books LIMIT 10;"
```

---

## 🧪 API接口测试

### 测试榜单接口（无需登录）

```bash
# 点赞周榜
curl http://localhost:8888/api/ranking/likes/week

# 点赞月榜
curl http://localhost:8888/api/ranking/likes/month

# 收藏周榜
curl http://localhost:8888/api/ranking/favorites/week

# 收藏月榜
curl http://localhost:8888/api/ranking/favorites/month
```

### 完整API测试脚本

```bash
cd /Users/dusong/GolandProjects/bookadmin
./test-api.sh
```

---

## 🐛 常见问题

### Q1: 前端启动失败 - "Missing script: dev"
**A**: 使用 `npm run serve` 而不是 `npm run dev`

### Q2: 后端启动失败 - "Redis连接失败"
**A**: 检查Redis是否启动
```bash
docker-compose ps
docker-compose logs redis
```

### Q3: 操作栏不显示
**A**: 
1. 确保鼠标悬停在图书卡片上
2. 尝试切换到轮播视图模式
3. 检查浏览器控制台是否有错误

### Q4: 点击按钮没反应
**A**:
1. 检查是否已登录
2. 查看浏览器控制台Network标签
3. 检查后端服务是否正常运行

### Q5: 计数不更新
**A**:
1. 刷新页面重新加载
2. 检查Redis是否正常运行
3. 查看后端日志

---

## 🔧 开发模式

### 前端热重载

```bash
cd web
npm run serve
# 修改代码后自动刷新浏览器
```

### 后端热重载（安装Air）

```bash
# 安装Air
go install github.com/cosmtrek/air@latest

# 启动热重载
cd /Users/dusong/GolandProjects/bookadmin
air
```

---

## 📊 性能监控

### Redis监控

```bash
# 实时监控Redis命令
docker exec -it bookadmin-redis redis-cli MONITOR

# 查看Redis信息
docker exec -i bookadmin-redis redis-cli INFO

# 查看连接数
docker exec -i bookadmin-redis redis-cli CLIENT LIST
```

### MySQL监控

```bash
# 查看进程列表
docker exec -i bookadmin-mysql mysql -uroot -proot -e "SHOW PROCESSLIST;"

# 查看慢查询
docker exec -i bookadmin-mysql mysql -uroot -proot -e "SHOW VARIABLES LIKE 'slow_query%';"
```

### Go服务监控

```bash
# 查看Go运行时信息
curl http://localhost:8888/debug/pprof/

# 查看goroutine
curl http://localhost:8888/debug/pprof/goroutine?debug=1
```

---

## 🛑 停止服务

### 停止前端

在前端终端按 `Ctrl + C`

### 停止后端

在后端终端按 `Ctrl + C`（会触发优雅关闭）

### 停止Docker

```bash
docker-compose down
```

---

## 🔄 重启服务

```bash
# 完全重启
docker-compose down
docker-compose up -d
./start-backend.sh
cd web && npm run serve
```

---

## 📝 日志查看

### 后端日志

后端日志直接显示在终端，或者：

```bash
# 如果使用nohup启动
tail -f nohup.out
```

### Docker日志

```bash
# MySQL日志
docker-compose logs -f mysql

# Redis日志
docker-compose logs -f redis
```

### 前端日志

- 浏览器控制台（F12）
- Network标签查看API请求
- Console标签查看错误信息

---

## 🎉 成功标志

✅ **后端运行正常**:
- 看到 "Worker池启动完成，共 5 个Worker"
- 看到 "服务器启动在端口: 8888"
- 没有 ERROR 日志

✅ **前端运行正常**:
- 看到 "App running at: http://localhost:8080"
- 浏览器可以正常打开页面
- 可以正常登录

✅ **功能运行正常**:
- 图书列表正常加载
- 鼠标悬停显示操作栏
- 点击按钮有动画效果
- 提示消息正常显示
- 计数实时更新

---

**最后更新**: 2025-11-11  
**系统版本**: v2.0 with Like/Favorite  
**状态**: ✅ 可用  

