#!/bin/bash
# 分布式追踪验证脚本：验证 X-Trace-ID 响应头与 Jaeger 可访问性
# 使用方式：./scripts/verify_tracing.sh
#
# 环境变量：
#   BASE_URL    - API 地址。Docker 全栈: http://localhost；本地 go run: http://localhost:8888
#   JAEGER_UI_URL - Jaeger UI 地址，默认 http://localhost:16686
#
# 前提：全栈已启动（docker compose up -d）或本地 go run main.go 且 config.yaml 中 tracing.enabled=true

set -e
BASE="${BASE_URL:-http://localhost}"
JAEGER_UI="${JAEGER_UI_URL:-http://localhost:16686}"

echo "=== 1. 检查 X-Trace-ID 响应头 ==="
RESPONSE_HEADERS=$(curl -sI "$BASE/api/healthz")
TRACE_ID=$(echo "$RESPONSE_HEADERS" | grep -i x-trace-id | cut -d: -f2 | tr -d ' \r')
TRACING_STATUS=$(echo "$RESPONSE_HEADERS" | grep -i x-tracing-status | cut -d: -f2 | tr -d ' \r')
if [ -n "$TRACE_ID" ]; then
  echo "✓ 获取到 X-Trace-ID: $TRACE_ID"
elif [ -n "$TRACING_STATUS" ]; then
  echo "✗ 未获取到 X-Trace-ID，X-Tracing-Status: $TRACING_STATUS"
  echo "  若为 disabled：请检查 docker-compose 中 TRACING_ENABLED=true 或 config.yaml 中 tracing.enabled"
  echo "  若为 enabled：otelgin 可能未正确注入 span，请查看 api 容器启动日志"
  exit 1
else
  echo "✗ 未获取到 X-Trace-ID，且无 X-Tracing-Status（请求可能未到达 Go 后端，检查 nginx 或 BASE_URL）"
  exit 1
fi

echo ""
echo "=== 2. 检查 Jaeger UI 可访问性 ==="
if curl -sf "$JAEGER_UI" > /dev/null; then
  echo "✓ Jaeger UI 可访问: $JAEGER_UI"
else
  echo "✗ Jaeger UI 不可访问，请确认 docker compose 中 jaeger 服务已启动"
fi

echo ""
echo "=== 3. 发起带业务的请求以生成完整链路 ==="
TOKEN=$(curl -s -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
  # 使用 GET 而非 HEAD，确保路由正确匹配（部分框架对 HEAD 支持不完整）
  USERINFO_HEADERS=$(curl -s -o /dev/null -D - -H "Authorization: Bearer $TOKEN" "$BASE/api/auth/userInfo")
  TRACE_ID2=$(echo "$USERINFO_HEADERS" | grep -i x-trace-id | cut -d: -f2 | tr -d ' \r')
  if [ -n "$TRACE_ID2" ]; then
    echo "✓ 业务请求 X-Trace-ID: $TRACE_ID2"
  else
    echo "⚠ 业务请求未返回 X-Trace-ID（检查返回状态: echo \"\$USERINFO_HEADERS\" | head -1）"
  fi
fi

echo ""
echo "验证完成。在 Jaeger UI ($JAEGER_UI) 中搜索 Service: bookadmin-api 可查看完整 trace"
