#!/bin/bash

# 测试登录并获取菜单

echo "1. 测试登录 (root用户)..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"root123"}')

echo "登录响应："
echo "$LOGIN_RESPONSE" | python -m json.tool

# 提取token
TOKEN=$(echo "$LOGIN_RESPONSE" | python -c "import sys, json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "无法获取token"
  exit 1
fi

echo ""
echo "2. 获取的Token: $TOKEN"
echo ""

echo "3. 测试获取用户信息..."
curl -s -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | python -m json.tool

echo ""
echo "4. 测试获取菜单..."
curl -s -X GET http://localhost:8080/api/auth/menus \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | python -m json.tool

echo ""
echo "5. 测试获取权限..."
curl -s -X GET http://localhost:8080/api/auth/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | python -m json.tool