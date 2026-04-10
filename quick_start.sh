#!/bin/bash

echo "========================================="
echo "   秒杀系统快速启动脚本"
echo "========================================="

PROJECT_DIR="/home/monody/project/Microsecond killing service"

echo -e "\n📦 步骤 1: 启动基础设施..."
cd "$PROJECT_DIR"
docker compose up -d

echo -e "\n⏳ 等待服务启动..."
sleep 5

echo -e "\n🔍 检查服务状态..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo -e "\n📊 步骤 2: 初始化 Redis 库存..."
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 3 2>/dev/null

echo -e "\n✅ 基础设施启动完成！"
echo -e "\n📝 接下来请在三个独立的终端中运行以下命令:"
echo -e "\n终端 1 - 用户服务:"
echo "  cd \"$PROJECT_DIR/user-main\" && /tmp/user-server -f etc/user.yaml"

echo -e "\n终端 2 - 秒杀服务:"
echo "  cd \"$PROJECT_DIR/seckill-main\" && /tmp/seckill-server -f etc/seckill.yaml"

echo -e "\n终端 3 - 网关服务:"
echo "  cd \"$PROJECT_DIR/gateway-main\" && /tmp/gateway-server -f etc/gateway.yaml"

echo -e "\n🧪 服务启动后，运行测试:"
echo "  cd \"$PROJECT_DIR\" && ./test_api.sh"

echo -e "\n📚 查看完整文档:"
echo "  cat \"$PROJECT_DIR/测试文档.md\""
