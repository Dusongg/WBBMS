#!/bin/bash

echo "=================================="
echo "🧪 API 接口测试脚本"
echo "=================================="
echo ""

BASE_URL="http://localhost:8888/api"
TOKEN=""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. 登录获取 Token
echo "📝 步骤1: 登录获取 Token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ 登录失败，请检查用户名密码${NC}"
    echo "响应: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ 登录成功${NC}"
echo "Token: ${TOKEN:0:20}..."
echo ""

# 2. 测试点赞功能
echo "📝 步骤2: 测试点赞功能..."
echo ""

# 2.1 点赞图书ID=1
echo "  2.1 点赞图书 (ID=1)..."
LIKE_RESPONSE=$(curl -s -X POST "$BASE_URL/like/toggle/1" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $LIKE_RESPONSE"
echo ""

# 2.2 查询点赞状态
echo "  2.2 查询点赞状态..."
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/like/status/1" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $STATUS_RESPONSE"
echo ""

# 2.3 批量查询点赞状态
echo "  2.3 批量查询点赞状态 (ID=1,2,3)..."
BATCH_RESPONSE=$(curl -s -X GET "$BASE_URL/like/batch-status?bookIds=1,2,3" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $BATCH_RESPONSE"
echo ""

# 3. 测试收藏功能
echo "📝 步骤3: 测试收藏功能..."
echo ""

# 3.1 收藏图书ID=1
echo "  3.1 收藏图书 (ID=1)..."
FAV_RESPONSE=$(curl -s -X POST "$BASE_URL/favorite/toggle/1" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $FAV_RESPONSE"
echo ""

# 3.2 查询收藏状态
echo "  3.2 查询收藏状态..."
FAV_STATUS=$(curl -s -X GET "$BASE_URL/favorite/status/1" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $FAV_STATUS"
echo ""

# 4. 测试榜单功能
echo "📝 步骤4: 测试榜单功能（无需登录）..."
echo ""

# 4.1 点赞周榜
echo "  4.1 点赞周榜 Top 10..."
WEEK_RANK=$(curl -s -X GET "$BASE_URL/ranking/likes/week?limit=10")
echo "  响应: $WEEK_RANK"
echo ""

# 4.2 点赞月榜
echo "  4.2 点赞月榜 Top 10..."
MONTH_RANK=$(curl -s -X GET "$BASE_URL/ranking/likes/month?limit=10")
echo "  响应: $MONTH_RANK"
echo ""

# 4.3 收藏周榜
echo "  4.3 收藏周榜 Top 10..."
FAV_WEEK_RANK=$(curl -s -X GET "$BASE_URL/ranking/favorites/week?limit=10")
echo "  响应: $FAV_WEEK_RANK"
echo ""

# 5. 查看我的点赞列表
echo "📝 步骤5: 查看我的点赞列表..."
MY_LIKES=$(curl -s -X GET "$BASE_URL/like/list?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $MY_LIKES"
echo ""

# 6. 查看我的收藏列表
echo "📝 步骤6: 查看我的收藏列表..."
MY_FAVS=$(curl -s -X GET "$BASE_URL/favorite/list?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN")
echo "  响应: $MY_FAVS"
echo ""

echo "=================================="
echo -e "${GREEN}✅ 测试完成！${NC}"
echo "=================================="
echo ""
echo "💡 提示："
echo "  - 如果响应中有 \"code\":200 表示成功"
echo "  - 如果响应中有 \"code\":401 表示未登录或Token过期"
echo "  - 如果响应中有 \"code\":500 表示服务器错误"
echo ""
echo "🔍 查看 Redis 数据："
echo "  docker exec -i bookadmin-redis redis-cli keys '*'"
echo ""
echo "🔍 查看 MySQL 数据："
echo "  docker exec -i bookadmin-mysql mysql -uroot -proot bookadmin -e 'SELECT * FROM book_likes;'"
echo ""

