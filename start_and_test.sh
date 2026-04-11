#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_DIR="/home/monody/project/SecKill"
GATEWAY_URL="http://localhost:8998"

USER_PORT=8669
SECKILL_PORT=8002
GATEWAY_PORT=8998

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未安装"
        exit 1
    fi
}

kill_port_process() {
    local port=$1
    local service_name=$2
    
    if lsof -i:$port &> /dev/null; then
        print_warning "端口 $port 被占用，正在清理..."
        lsof -ti:$port | xargs -r kill -9 2>/dev/null || true
        sleep 1
        print_success "端口 $port 已清理"
    else
        print_info "端口 $port 可用"
    fi
}

check_and_clean_ports() {
    print_header "检查并清理端口"
    
    kill_port_process $USER_PORT "user RPC"
    kill_port_process $SECKILL_PORT "seckill RPC"
    kill_port_process $GATEWAY_PORT "gateway HTTP"
}

wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local max_attempts=30
    local attempt=1
    
    while ! nc -z $host $port 2>/dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            print_error "$service 启动超时"
            return 1
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    print_success "$service 已就绪 ($host:$port)"
    return 0
}

start_infrastructure() {
    print_header "启动基础设施服务"
    
    cd "$PROJECT_DIR"
    
    if docker ps --format '{{.Names}}' | grep -q "mks-etcd"; then
        print_info "基础设施服务已在运行中"
    else
        docker compose up -d
        sleep 3
    fi
    
    wait_for_service localhost 20001 "etcd"
    wait_for_service localhost 3307 "MySQL"
    wait_for_service localhost 6379 "Redis"
    wait_for_service localhost 9092 "Kafka"
}

init_go_modules() {
    print_header "初始化Go模块依赖"
    
    cd "$PROJECT_DIR/user-main"
    if [ ! -f "go.sum" ]; then
        go mod tidy
    fi
    print_success "user-main 依赖已就绪"
    
    cd "$PROJECT_DIR/seckill-main"
    if [ ! -f "go.sum" ]; then
        go mod tidy
    fi
    print_success "seckill-main 依赖已就绪"
    
    cd "$PROJECT_DIR/gateway-main"
    if [ ! -f "go.sum" ]; then
        go mod tidy
    fi
    print_success "gateway-main 依赖已就绪"
}

start_services() {
    print_header "启动微服务"
    
    check_and_clean_ports
    
    cd "$PROJECT_DIR/user-main"
    go run ./cmd/user -f etc/user.yaml > /tmp/user.log 2>&1 &
    USER_PID=$!
    print_info "启动 user RPC 服务 (PID: $USER_PID)"
    
    cd "$PROJECT_DIR/seckill-main"
    go run ./cmd/sec_kill -f etc/seckill.yaml > /tmp/seckill.log 2>&1 &
    SECKILL_PID=$!
    print_info "启动 seckill RPC 服务 (PID: $SECKILL_PID)"
    
    cd "$PROJECT_DIR/gateway-main"
    go run ./cmd/gateway -f etc/gateway.yaml > /tmp/gateway.log 2>&1 &
    GATEWAY_PID=$!
    print_info "启动 gateway HTTP 服务 (PID: $GATEWAY_PID)"
    
    sleep 5
    
    wait_for_service 127.0.0.1 $USER_PORT "user RPC"
    wait_for_service 127.0.0.1 $SECKILL_PORT "seckill RPC"
    wait_for_service 0.0.0.0 $GATEWAY_PORT "gateway HTTP"
    
    echo ""
    print_info "服务进程 PID:"
    echo "  user:     $USER_PID"
    echo "  seckill:  $SECKILL_PID"
    echo "  gateway:  $GATEWAY_PID"
}

setup_redis_stock() {
    print_header "设置Redis库存数据"
    
    docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 10 2>/dev/null
    docker exec mks-redis redis-cli -a 123456 SET "SK:Limit1" 5 2>/dev/null
    
    docker exec mks-redis redis-cli -a 123456 DEL "SK:UserGoodsSecNum:1:1" 2>/dev/null
    docker exec mks-redis redis-cli -a 123456 DEL "SK:UserSecKilledNum:1:1" 2>/dev/null
    
    print_success "Redis库存已设置 (SK:Stock:1 = 10, SK:Limit1 = 5)"
    print_success "已清理用户购买记录"
}

test_api() {
    print_header "测试API接口"
    
    local pass_count=0
    local fail_count=0
    
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 1: 登录接口 POST /login${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    LOGIN_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"123321"}')
    
    CODE=$(echo $LOGIN_RESPONSE | grep -o '"code":[0-9]*' | cut -d':' -f2)
    
    if [ "$CODE" == "200" ]; then
        TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        print_success "登录成功"
        echo "  Token: ${TOKEN:0:50}..."
        ((pass_count++))
    else
        print_error "登录失败: $LOGIN_RESPONSE"
        ((fail_count++))
        return 1
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 2: 获取欢迎信息 GET /get_user_info${NC}"
    echo -e "${YELLOW}注意: 需要使用 Authorization: Bearer <token> 格式${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    WELCOME_INFO=$(curl -s "$GATEWAY_URL/get_user_info" \
        -H "Authorization: Bearer $TOKEN")
    
    WELCOME=$(echo $WELCOME_INFO | grep -o '"welcome":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$WELCOME" ]; then
        print_success "获取欢迎信息成功"
        echo "  欢迎信息: $WELCOME"
        ((pass_count++))
    else
        print_error "获取欢迎信息失败: $WELCOME_INFO"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 3: 获取用户详细信息 GET /bitstorm/get_user_info${NC}"
    echo -e "${YELLOW}注意: 需要使用 Authorization: Bearer <token> 格式${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    USER_INFO=$(curl -s "$GATEWAY_URL/bitstorm/get_user_info" \
        -H "Authorization: Bearer $TOKEN")
    
    CODE=$(echo $USER_INFO | grep -o '"code":[0-9]*' | cut -d':' -f2)
    
    if [ "$CODE" == "0" ]; then
        print_success "获取用户信息成功"
        USERNAME=$(echo $USER_INFO | grep -o '"userName":"[^"]*"' | cut -d'"' -f4)
        echo "  用户名: $USERNAME"
        ((pass_count++))
    else
        print_error "获取用户信息失败: $USER_INFO"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 4: 按用户名获取用户信息 GET /bitstorm/get_user_info_by_name${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    USER_BY_NAME=$(curl -s "$GATEWAY_URL/bitstorm/get_user_info_by_name?user_name=admin" \
        -H "Authorization: Bearer $TOKEN")
    
    CODE=$(echo $USER_BY_NAME | grep -o '"code":[0-9]*' | cut -d':' -f2)
    
    if [ "$CODE" == "0" ]; then
        print_success "按用户名获取用户信息成功"
        echo "  响应: $USER_BY_NAME"
        ((pass_count++))
    else
        print_error "按用户名获取用户信息失败: $USER_BY_NAME"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 5: 秒杀v1接口 POST /bitstorm/v1/sec_kill${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    SECKILL_V1=$(curl -s -X POST "$GATEWAY_URL/bitstorm/v1/sec_kill" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"goodsNum":"abc123","num":1}')
    
    ORDER_NUM=$(echo $SECKILL_V1 | grep -o '"orderNum":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$ORDER_NUM" ]; then
        print_success "秒杀v1成功"
        echo "  订单号: $ORDER_NUM"
        ((pass_count++))
    else
        print_error "秒杀v1失败: $SECKILL_V1"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 6: 秒杀v2接口 POST /bitstorm/v2/sec_kill${NC}"
    echo -e "${YELLOW}注意: 需要在Redis中设置库存 SK:Stock:1${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    SECKILL_V2=$(curl -s -X POST "$GATEWAY_URL/bitstorm/v2/sec_kill" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"goodsNum":"abc123","num":1}')
    
    ORDER_NUM=$(echo $SECKILL_V2 | grep -o '"orderNum":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$ORDER_NUM" ]; then
        print_success "秒杀v2成功"
        echo "  订单号: $ORDER_NUM"
        ((pass_count++))
    else
        print_error "秒杀v2失败: $SECKILL_V2"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 7: 秒杀v3接口 POST /bitstorm/v3/sec_kill${NC}"
    echo -e "${YELLOW}注意: 需要在Redis中设置库存 SK:Stock:1${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    SECKILL_V3=$(curl -s -X POST "$GATEWAY_URL/bitstorm/v3/sec_kill" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"goodsNum":"abc123","num":1}')
    
    SEC_NUM=$(echo $SECKILL_V3 | grep -o '"secNum":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$SEC_NUM" ]; then
        print_success "秒杀v3成功"
        echo "  秒杀号: $SEC_NUM"
        ((pass_count++))
    else
        print_error "秒杀v3失败: $SECKILL_V3"
        ((fail_count++))
    fi
    
    echo ""
    echo -e "${YELLOW}----------------------------------------${NC}"
    echo -e "${YELLOW}测试 8: 查询秒杀状态 GET /bitstorm/v3/get_sec_kill_info${NC}"
    echo -e "${YELLOW}----------------------------------------${NC}"
    
    if [ -n "$SEC_NUM" ]; then
        SECKILL_INFO=$(curl -s "$GATEWAY_URL/bitstorm/v3/get_sec_kill_info?sec_num=$SEC_NUM" \
            -H "Authorization: Bearer $TOKEN")
        
        STATUS=$(echo $SECKILL_INFO | grep -o '"status":[0-9]*' | cut -d':' -f2)
        
        if [ -n "$STATUS" ]; then
            print_success "查询秒杀状态成功"
            echo "  状态: $STATUS"
            echo "  响应: $SECKILL_INFO"
            ((pass_count++))
        else
            print_error "查询秒杀状态失败: $SECKILL_INFO"
            ((fail_count++))
        fi
    else
        print_error "跳过: 无有效秒杀号"
        ((fail_count++))
    fi
    
    echo ""
    print_header "测试结果汇总"
    
    echo -e "  ${GREEN}通过: $pass_count${NC}"
    echo -e "  ${RED}失败: $fail_count${NC}"
    echo ""
    
    if [ $fail_count -eq 0 ]; then
        print_success "所有测试通过！"
        return 0
    else
        print_error "部分测试失败"
        return 1
    fi
}

stop_services() {
    print_header "停止服务"
    
    print_info "清理端口 $USER_PORT..."
    lsof -ti:$USER_PORT | xargs -r kill -9 2>/dev/null || true
    
    print_info "清理端口 $SECKILL_PORT..."
    lsof -ti:$SECKILL_PORT | xargs -r kill -9 2>/dev/null || true
    
    print_info "清理端口 $GATEWAY_PORT..."
    lsof -ti:$GATEWAY_PORT | xargs -r kill -9 2>/dev/null || true
    
    print_success "微服务已停止"
}

show_logs() {
    echo ""
    print_info "日志文件位置:"
    echo "  user:     /tmp/user.log"
    echo "  seckill:  /tmp/seckill.log"
    echo "  gateway:  /tmp/gateway.log"
}

show_usage() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║        SecKill 项目一键启动测试脚本 - 帮助信息             ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}用法:${NC}"
    echo "  $0 [命令]"
    echo ""
    echo -e "${GREEN}命令列表:${NC}"
    echo -e "  ${YELLOW}(无参数)${NC}    完整流程：启动基础设施 → 初始化依赖 → 启动服务 → 测试API"
    echo -e "  ${YELLOW}start${NC}       仅启动服务（基础设施 + 微服务 + Redis库存设置）"
    echo -e "  ${YELLOW}test${NC}        仅测试API接口（需先运行 start 或无参数启动）"
    echo -e "  ${YELLOW}stop${NC}        停止所有微服务进程"
    echo -e "  ${YELLOW}clean${NC}       清理端口占用（8669, 8002, 8998）"
    echo -e "  ${YELLOW}help${NC}        显示此帮助信息"
    echo ""
    echo -e "${GREEN}测试的API接口:${NC}"
    echo "  1. POST /login                           登录获取Token"
    echo "  2. GET  /get_user_info                   获取欢迎信息"
    echo "  3. GET  /bitstorm/get_user_info          获取用户详细信息"
    echo "  4. GET  /bitstorm/get_user_info_by_name  按用户名查询"
    echo "  5. POST /bitstorm/v1/sec_kill            秒杀v1（数据库扣减）"
    echo "  6. POST /bitstorm/v2/sec_kill            秒杀v2（Redis扣减）"
    echo "  7. POST /bitstorm/v3/sec_kill            秒杀v3（异步消息队列）"
    echo "  8. GET  /bitstorm/v3/get_sec_kill_info   查询秒杀状态"
    echo ""
    echo -e "${GREEN}服务端口:${NC}"
    echo "  user RPC      : 8669"
    echo "  seckill RPC   : 8002"
    echo "  gateway HTTP  : 8998"
    echo "  etcd          : 20001"
    echo "  MySQL         : 3307"
    echo "  Redis         : 6379"
    echo "  Kafka         : 9092"
    echo ""
    echo -e "${GREEN}注意事项:${NC}"
    echo "  • 秒杀v2/v3 需要在Redis中设置库存: SK:Stock:1"
    echo "  • 认证接口需要使用 Header: Authorization: Bearer <token>"
    echo "  • 日志文件位置: /tmp/user.log, /tmp/seckill.log, /tmp/gateway.log"
    echo ""
    echo -e "${GREEN}示例:${NC}"
    echo "  $0              # 完整启动并测试"
    echo "  $0 start        # 仅启动服务"
    echo "  $0 test         # 仅测试API"
    echo "  $0 stop         # 停止服务"
    echo ""
}

main() {
    print_header "SecKill 项目一键启动测试脚本"
    
    check_command docker
    check_command go
    check_command curl
    check_command nc
    check_command lsof
    
    start_infrastructure
    init_go_modules
    start_services
    setup_redis_stock
    
    if test_api; then
        show_logs
        print_success "测试完成！服务仍在运行中"
        echo ""
        print_info "如需停止服务，请运行: $0 stop"
    else
        show_logs
        print_error "测试失败，请检查日志"
        exit 1
    fi
}

case "${1:-}" in
    start)
        start_infrastructure
        init_go_modules
        start_services
        setup_redis_stock
        show_logs
        ;;
    test)
        test_api
        ;;
    stop)
        stop_services
        ;;
    clean)
        check_and_clean_ports
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        main
        ;;
esac
