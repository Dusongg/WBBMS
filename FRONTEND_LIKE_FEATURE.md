# 前端点赞/收藏功能说明

## ✅ 已完成功能

### 📁 新增文件 (3个)
```
web/src/api/like.js         - 点赞API接口
web/src/api/favorite.js     - 收藏API接口
web/src/api/ranking.js      - 榜单API接口
```

### ✏️ 修改文件 (1个)
```
web/src/views/BookList.vue  - 图书列表页面
  - 添加点赞/收藏状态管理
  - 实现批量查询功能
  - 添加点赞/收藏按钮UI
  - 添加动画效果
```

---

## 🎨 功能特性

### 1. 智能状态管理
- **批量查询**：获取图书列表后，一次性批量查询所有图书的点赞/收藏状态
- **实时更新**：点击后立即更新UI状态，无需刷新页面
- **防重复点击**：添加loading状态，防止重复提交

### 2. 精美UI设计
- **悬浮显示**：鼠标悬停或中心卡片自动显示操作栏
- **渐变背景**：操作栏使用半透明渐变背景，不遮挡封面
- **平滑动画**：
  - 操作栏淡入淡出
  - 按钮悬浮效果
  - 点赞弹跳动画 ❤️
  - 收藏旋转动画 ⭐

### 3. 交互体验
- **视觉反馈**：
  - 未点赞：白色心形 🤍 + 计数
  - 已点赞：红色心形 ❤️ + 弹跳动画
  - 未收藏：空心星星 ☆ + 计数
  - 已收藏：实心星星 ⭐ + 旋转动画
- **即时提示**：
  - 点赞成功：提示 "点赞成功 ❤️"
  - 取消点赞：提示 "已取消点赞"
  - 收藏成功：提示 "收藏成功 ⭐"
  - 取消收藏：提示 "已取消收藏"

---

## 📊 数据流向

```
1. 用户进入图书列表
   ↓
2. fetchBookList() 获取图书数据
   ↓
3. fetchLikeAndFavoriteStatus() 批量查询状态
   ├─ batchGetLikeStatus(bookIds)
   └─ batchGetFavoriteStatus(bookIds)
   ↓
4. 更新 likeStatus 和 favoriteStatus
   ↓
5. UI自动渲染按钮状态

点击按钮时：
1. handleToggleLike/handleToggleFavorite
   ↓
2. 调用后端API切换状态
   ↓
3. 更新本地状态
   ↓
4. 显示成功提示
   ↓
5. 播放动画效果
```

---

## 🎯 使用方法

### 启动前端

```bash
cd /Users/dusong/GolandProjects/bookadmin/web
npm install
npm run dev
```

### 测试功能

1. **登录系统**：使用 admin/admin123
2. **进入图书列表**：点击顶部"图书列表"
3. **鼠标悬停**：在任意图书卡片上悬停
4. **查看操作栏**：底部会出现点赞❤️和收藏⭐按钮
5. **点击测试**：
   - 点击❤️按钮：看到弹跳动画和红色填充
   - 点击⭐按钮：看到旋转动画和黄色填充
   - 再次点击：取消并恢复原状

---

## 🔧 API调用示例

```javascript
// 批量查询点赞状态
const response = await batchGetLikeStatus([1, 2, 3, 4, 5])
// 返回: [
//   { book_id: 1, is_liked: true, like_count: 10 },
//   { book_id: 2, is_liked: false, like_count: 5 },
//   ...
// ]

// 切换点赞
const response = await toggleLike(1)
// 返回: { is_liked: true, like_count: 11 }

// 切换收藏
const response = await toggleFavorite(1)
// 返回: { is_favorited: true, favorite_count: 8 }
```

---

## 🎨 样式定制

### 自定义颜色

```css
/* 点赞按钮颜色 */
.like-btn {
  color: #ef4444;  /* 红色 */
}

/* 收藏按钮颜色 */
.favorite-btn {
  color: #f59e0b;  /* 黄色 */
}

/* 操作栏背景 */
.book-action-bar {
  background: linear-gradient(to top, rgba(0, 0, 0, 0.7), rgba(0, 0, 0, 0));
}
```

### 自定义动画

```css
/* 点赞弹跳动画 */
@keyframes like-bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.3); }
}

/* 收藏旋转动画 */
@keyframes favorite-rotate {
  0%, 100% { transform: rotate(0deg) scale(1); }
  50% { transform: rotate(72deg) scale(1.3); }
}
```

---

## 🐛 常见问题

### Q1: 点击按钮后卡片也被点击了？
**A**: 已处理。使用 `@click.stop` 阻止事件冒泡。

### Q2: 重复点击会发送多次请求？
**A**: 已处理。使用 `actionLoading` 状态防止重复点击。

### Q3: 操作栏一直显示很突兀？
**A**: 已优化。仅在鼠标悬停或中心卡片时显示，平时自动隐藏。

### Q4: 动画效果太慢？
**A**: 可调整CSS transition时间：
```css
.book-action-bar {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);  /* 调小0.3s */
}
```

---

## 📱 响应式设计

操作栏在不同屏幕尺寸下自适应：
- **桌面**: 完整显示，悬停显示
- **平板**: 缩小按钮，点击显示
- **手机**: 简化UI，长按显示

---

## 🚀 下一步优化（可选）

1. **长按显示菜单**：长按卡片显示更多操作
2. **双击快速收藏**：双击图片直接收藏
3. **拖动收藏**：拖动图片到收藏区域
4. **批量操作**：选中多本书批量点赞/收藏
5. **分享功能**：分享喜欢的图书给其他用户

---

**开发完成时间**: 2025-11-11  
**版本**: v1.0  
**状态**: ✅ 已完成并可投入使用  

