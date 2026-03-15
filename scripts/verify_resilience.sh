#!/bin/bash
# 限流与熔断验证脚本
# 使用方式：./scripts/verify_resilience.sh
#
# 环境变量：
#   BASE_URL           - API 地址，默认 http://localhost
#   SKIP_RATE_LIMIT    - 设为 1 跳过限流验证
#   SKIP_CIRCUIT_BREAKER - 设为 1 跳过熔断验证
#   REDIS_CONTAINER    - Redis 容器名，默认 bookadmin-redis
#
# 限流验证：快速并发请求非健康检查路径，统计 200/429 比例，出现 429 即通过
# 熔断验证：停止 Redis → 触发熔断 → 业务接口返回 503 → 恢复 Redis → 等待半开
#
# 前提：全栈已启动（docker compose up -d）或本地 go run，且 config.yaml 中
#       resilience.rate-limit.enabled=true、resilience.circuit-breaker.enabled=true

set -e
BASE="${BASE_URL:-http://localhost}"
REDIS_CONTAINER="${REDIS_CONTAINER:-bookadmin-redis}"

# 限流测试使用的路径（需为非 skip 路径：非 /healthz、/readyz、/metrics、/api/healthz、/api/readyz）
RATE_LIMIT_PATH="${RATE_LIMIT_PATH:-$BASE/api/category/getCategoryList}"
# 熔断测试时用于触发失败的探针路径（会调用 Redis Ping，需与 nginx 代理一致）
READYZ_PATH="${READYZ_PATH:-$BASE/api/readyz}"
# 熔断打开后应返回 503 的业务路径
BUSINESS_PATH="${BUSINESS_PATH:-$BASE/api/category/getCategoryList}"

echo "=== 限流与熔断验证 ==="
echo "BASE_URL=$BASE"
echo ""

# -----------------------------------------------------------------------------
# 1. 限流验证
# -----------------------------------------------------------------------------
if [ "${SKIP_RATE_LIMIT}" != "1" ]; then
  echo "=== 1. 限流验证 ==="
  echo "快速并发请求 $RATE_LIMIT_PATH（约 250 次），统计 HTTP 状态码："
  echo "（若未出现 429，可尝试将 config.yaml 中 rate-limit.rps 调小如 10 以便验证）"
  echo ""

  RESULT=$(seq 1 250 2>/dev/null | xargs -P 30 -I {} curl -s -o /dev/null -w "%{http_code}\n" "$RATE_LIMIT_PATH" 2>/dev/null || true)
  if [ -z "$RESULT" ]; then
    # fallback when xargs/seq differ (e.g. macOS)
    COUNTS=""
    for i in $(seq 1 250 2>/dev/null || jot 250 1 250 2>/dev/null || echo "1 2 3"); do
      CODE=$(curl -s -o /dev/null -w "%{http_code}" "$RATE_LIMIT_PATH" 2>/dev/null || echo "000")
      COUNTS="${COUNTS}${CODE}"$'\n'
    done
    RESULT=$(echo "$COUNTS" | grep -v '^$')
  fi

  STATS=$(echo "$RESULT" | sort | uniq -c | sort -rn)
  echo "$STATS"
  echo ""

  if echo "$STATS" | grep -q 429; then
    echo "✓ 出现 429，限流生效"
  else
    echo "⚠ 未出现 429。若限流已启用，可能因 RPS/Burst 较高或并发不足。"
    echo "  可临时将 config 中 rate-limit.rps 设为 10、burst 设为 5 再试"
  fi
  echo ""
else
  echo "=== 1. 限流验证（已跳过） ==="
  echo ""
fi

# -----------------------------------------------------------------------------
# 2. 熔断验证
# -----------------------------------------------------------------------------
if [ "${SKIP_CIRCUIT_BREAKER}" != "1" ]; then
  echo "=== 2. 熔断验证 ==="

  if ! command -v docker &>/dev/null; then
    echo "⚠ 未检测到 docker，跳过熔断验证（需停止 Redis 容器以模拟故障）"
  elif ! docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${REDIS_CONTAINER}$"; then
    echo "⚠ 未找到 Redis 容器 '$REDIS_CONTAINER'，跳过熔断验证"
  else
    echo "2.1 停止 Redis 容器 ($REDIS_CONTAINER)..."
    docker stop "$REDIS_CONTAINER" 2>/dev/null || true
    sleep 2

    echo "2.2 连续请求 $READYZ_PATH 以触发熔断（失败阈值默认 5）..."
    for i in $(seq 1 8); do
      CODE=$(curl -s -o /dev/null -w "%{http_code}" "$READYZ_PATH" 2>/dev/null || echo "000")
      echo "  请求 $i: HTTP $CODE"
    done

    echo ""
    echo "2.3 检查业务接口是否因熔断返回 503..."
    CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BUSINESS_PATH" 2>/dev/null || echo "000")
    if [ "$CODE" = "503" ]; then
      echo "✓ 熔断已打开，业务接口返回 503"
    else
      echo "  当前返回 HTTP $CODE（若熔断未打开，可能为 200/500 等）"
    fi

    echo ""
    echo "2.4 恢复 Redis 容器..."
    docker start "$REDIS_CONTAINER" 2>/dev/null || true
    echo "  等待 5 秒供 Redis 就绪..."
    sleep 5

    echo ""
    echo "2.5 熔断超时约 30 秒后进入半开，可稍后再次请求验证恢复。"
    echo "  当前再试一次 $BUSINESS_PATH："
    CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BUSINESS_PATH" 2>/dev/null || echo "000")
    echo "  HTTP $CODE"
  fi
  echo ""
else
  echo "=== 2. 熔断验证（已跳过） ==="
  echo ""
fi

echo "验证完成"
