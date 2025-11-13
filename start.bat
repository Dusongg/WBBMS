@echo off

echo 正在启动后端服务...
start cmd /k "go run main.go"

timeout /t 3 /nobreak >nul

echo 正在启动前端服务...
cd web
start cmd /k "npm run serve"

