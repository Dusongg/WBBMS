#!/bin/bash
# 无状态验证脚本：验证 API 多实例负载均衡与跨实例 Token 有效性
# 使用方式：在项目根目录执行 ./scripts/verify_stateless.sh
# 前提：docker compose --profile app up -d 已启动

set -e
BASE="${BASE_URL:-http://localhost}"

echo "=== 1. 负载均衡验证 ==="
echo "连续 12 次请求 /api/healthz，统计各实例命中次数："
for i in $(seq 1 12); do
  curl -sI "$BASE/api/healthz" | grep -i x-instance-id || true
done | sort | uniq -c

echo ""
echo "=== 2. 无状态验证（同一 Token 跨实例均有效） ==="
TOKEN=$(curl -s -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取 Token"
  exit 1
fi

FAIL=0
for i in $(seq 1 10); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/api/auth/userInfo")
  if [ "$CODE" != "200" ]; then
    echo "请求 $i 失败: HTTP $CODE"
    FAIL=1
  fi
done

if [ $FAIL -eq 0 ]; then
  echo "全部 10 次请求均返回 200，无状态验证通过"
else
  echo "存在失败请求，无状态验证未通过"
  exit 1
fi

echo ""
echo "=== 3. 认证请求的实例分布 ==="
for i in $(seq 1 6); do
  curl -sI -H "Authorization: Bearer $TOKEN" "$BASE/api/auth/userInfo" | grep -i x-instance-id || true
done | sort | uniq -c

echo ""
echo "验证完成"
