#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
START_SCRIPT="${PROJECT_DIR}/start_and_test.sh"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

GATEWAY_URL="${GATEWAY_URL:-http://127.0.0.1:8998}"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3307}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-123456}"
MYSQL_DB="${MYSQL_DB:-bitstorm}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-123456}"

REPORT_DIR="${PROJECT_DIR}/tmp/e2e_reports"
REPORT_FILE="${REPORT_DIR}/report_$(date +%Y%m%d_%H%M%S).txt"

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
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

check_prerequisites() {
    print_header "检查前置条件"

    local missing=0

    for cmd in curl docker nc lsof; do
        if ! command -v "${cmd}" &>/dev/null; then
            print_error "缺少命令: ${cmd}"
            ((missing++))
        else
            print_success "命令已安装: ${cmd}"
        fi
    done

    if [[ "${missing}" -gt 0 ]]; then
        print_error "请安装缺失的命令后重试"
        return 1
    fi

    return 0
}

check_services_running() {
    print_header "检查服务状态"

    local services_ok=1

    if ! nc -z "${MYSQL_HOST}" "${MYSQL_PORT}" 2>/dev/null; then
        print_error "MySQL 未运行 (${MYSQL_HOST}:${MYSQL_PORT})"
        services_ok=0
    else
        print_success "MySQL 已运行"
    fi

    if ! nc -z "${REDIS_HOST}" "${REDIS_PORT}" 2>/dev/null; then
        print_error "Redis 未运行 (${REDIS_HOST}:${REDIS_PORT})"
        services_ok=0
    else
        print_success "Redis 已运行"
    fi

    if ! curl -sS "${GATEWAY_URL}/health" >/dev/null 2>&1; then
        print_error "Gateway 未运行 (${GATEWAY_URL})"
        services_ok=0
    else
        print_success "Gateway 已运行"
    fi

    if [[ "${services_ok}" -eq 0 ]]; then
        print_warning "部分服务未运行，请先启动服务"
        return 1
    fi

    return 0
}

start_services() {
    print_header "启动服务"

    if [[ ! -f "${START_SCRIPT}" ]]; then
        print_error "启动脚本不存在: ${START_SCRIPT}"
        return 1
    fi

    print_info "调用启动脚本: ${START_SCRIPT} start"

    if bash "${START_SCRIPT}" start; then
        print_success "服务启动成功"
        return 0
    else
        print_error "服务启动失败"
        return 1
    fi
}

stop_services() {
    print_header "停止服务"

    if [[ ! -f "${START_SCRIPT}" ]]; then
        print_error "启动脚本不存在: ${START_SCRIPT}"
        return 1
    fi

    print_info "调用启动脚本: ${START_SCRIPT} stop"

    if bash "${START_SCRIPT}" stop; then
        print_success "服务已停止"
        return 0
    else
        print_warning "停止服务时出现警告"
        return 0
    fi
}

clean_environment() {
    print_header "清理测试环境"

    print_info "清理端口..."
    bash "${START_SCRIPT}" clean 2>/dev/null || true

    print_info "清理临时文件..."
    rm -rf "${PROJECT_DIR}/tmp/e2e_reports"/* 2>/dev/null || true
    rm -rf /tmp/e2e_test_logs/* 2>/dev/null || true

    print_success "环境清理完成"
}

run_tests() {
    local test_type="${1:-all}"

    print_header "执行端到端测试"

    mkdir -p "${REPORT_DIR}"

    export GATEWAY_URL
    export MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DB
    export REDIS_HOST REDIS_PORT REDIS_PASSWORD

    local test_script="${SCRIPT_DIR}/test_cases.sh"

    if [[ ! -f "${test_script}" ]]; then
        print_error "测试脚本不存在: ${test_script}"
        return 1
    fi

    print_info "测试类型: ${test_type}"
    print_info "报告目录: ${REPORT_DIR}"

    local exit_code=0

    case "${test_type}" in
        all)
            print_info "运行所有测试用例..."
            {
                echo "========================================"
                echo "端到端集成测试报告"
                echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
                echo "========================================"
                echo ""
                source "${test_script}" && run_all_tests
            } | tee "${REPORT_FILE}"
            exit_code=${PIPESTATUS[0]}
            ;;
        login)
            source "${test_script}" && test_login
            exit_code=$?
            ;;
        v1)
            source "${test_script}" && test_v1_flow
            exit_code=$?
            ;;
        v2)
            source "${test_script}" && test_v2_flow
            exit_code=$?
            ;;
        v3)
            source "${test_script}" && test_v3_flow
            exit_code=$?
            ;;
        concurrent)
            source "${test_script}" && test_concurrent_seckill
            exit_code=$?
            ;;
        quota)
            source "${test_script}" && test_quota_limit
            exit_code=$?
            ;;
        stock)
            source "${test_script}" && test_stock_exhausted
            exit_code=$?
            ;;
        *)
            print_error "未知的测试类型: ${test_type}"
            show_usage
            return 1
            ;;
    esac

    echo ""
    if [[ "${exit_code}" -eq 0 ]]; then
        print_success "测试完成"
        print_info "测试报告: ${REPORT_FILE}"
    else
        print_error "测试失败"
        print_info "测试报告: ${REPORT_FILE}"
    fi

    return ${exit_code}
}

show_usage() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           SecKill 端到端集成测试脚本 - 帮助信息            ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}用法:${NC}"
    echo "  $0 [命令] [测试类型]"
    echo ""
    echo -e "${GREEN}命令列表:${NC}"
    echo -e "  ${YELLOW}start${NC}       启动所有服务（基础设施 + 微服务）"
    echo -e "  ${YELLOW}test${NC}        运行测试用例（需先启动服务）"
    echo -e "  ${YELLOW}stop${NC}        停止所有服务"
    echo -e "  ${YELLOW}clean${NC}       清理测试环境"
    echo -e "  ${YELLOW}all${NC}         完整流程：启动服务 → 运行测试 → 停止服务"
    echo -e "  ${YELLOW}help${NC}        显示此帮助信息"
    echo ""
    echo -e "${GREEN}测试类型 (用于 test 命令):${NC}"
    echo -e "  ${YELLOW}all${NC}         运行所有测试用例 (默认)"
    echo -e "  ${YELLOW}login${NC}       测试用户登录"
    echo -e "  ${YELLOW}v1${NC}          测试 V1 完整流程 (数据库扣减)"
    echo -e "  ${YELLOW}v2${NC}          测试 V2 完整流程 (Redis 扣减)"
    echo -e "  ${YELLOW}v3${NC}          测试 V3 完整流程 (异步消息队列)"
    echo -e "  ${YELLOW}concurrent${NC}  测试并发秒杀"
    echo -e "  ${YELLOW}quota${NC}       测试限购功能"
    echo -e "  ${YELLOW}stock${NC}       测试库存耗尽"
    echo ""
    echo -e "${GREEN}环境变量:${NC}"
    echo "  GATEWAY_URL    Gateway 地址 (默认: http://127.0.0.1:8998)"
    echo "  MYSQL_HOST     MySQL 主机 (默认: 127.0.0.1)"
    echo "  MYSQL_PORT     MySQL 端口 (默认: 3307)"
    echo "  REDIS_HOST     Redis 主机 (默认: 127.0.0.1)"
    echo "  REDIS_PORT     Redis 端口 (默认: 6379)"
    echo ""
    echo -e "${GREEN}示例:${NC}"
    echo "  $0                          # 完整流程：启动、测试、停止"
    echo "  $0 start                    # 仅启动服务"
    echo "  $0 test                     # 运行所有测试用例"
    echo "  $0 test v1                  # 仅测试 V1 流程"
    echo "  $0 test concurrent          # 仅测试并发秒杀"
    echo "  $0 stop                     # 停止服务"
    echo ""
    echo -e "${GREEN}关键断言:${NC}"
    echo "  • 库存不超卖: 成功订单数 ≤ 初始库存"
    echo "  • 限购不超限: 同一用户成功秒杀次数 ≤ 限购数量"
    echo "  • 响应码正确: HTTP 200 且业务 code = 0"
    echo "  • 订单数量正确: 数据库订单数与成功秒杀数一致"
    echo ""
}

main() {
    print_header "SecKill 端到端集成测试"

    check_prerequisites || exit 1

    local command="${1:-all}"
    local test_type="${2:-all}"

    case "${command}" in
        start)
            start_services
            ;;
        test)
            check_services_running || {
                print_warning "服务未运行，尝试启动..."
                start_services || exit 1
                sleep 5
            }
            run_tests "${test_type}"
            ;;
        stop)
            stop_services
            ;;
        clean)
            clean_environment
            ;;
        all)
            start_services || exit 1
            sleep 5
            run_tests "all"
            local test_result=$?
            echo ""
            read -p "是否停止服务? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                stop_services
            fi
            exit ${test_result}
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "未知命令: ${command}"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
