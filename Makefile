.PHONY: test-unit test-integration smoke bench-smoke sync-config

test-unit:
	./scripts/test_unit.sh

test-integration:
	./scripts/test_integration.sh

smoke:
	./scripts/smoke.sh local

bench-smoke:
	./scripts/bench_smoke.sh

sync-config:
	./scripts/sync_runtime_configs.sh
