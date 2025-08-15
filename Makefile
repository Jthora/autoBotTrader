# Unified developer workflows

.PHONY: help all test api-test api-run api-run-file contracts-test frontend-dev lint ephem-generate tidy gas-snapshot gas-compare

help:
	@echo 'Common targets:'
	@echo '  make test              - run all tests (Go race + Cairo)'
	@echo '  make api-test          - Go API tests with race'
	@echo '  make api-run           - run API server (mock mode)'
	@echo '  make api-run-file      - run API server (file mode; needs EPHEM_TABLE_PATH)'
	@echo '  make contracts-test    - run Cairo contract tests (scarb test)'
	@echo '  make frontend-dev      - start frontend Vite dev server'
	@echo '  make gas-snapshot      - run cairo tests, extract gas snapshot'
	@echo '  make gas-compare       - compare snapshot against baselines'
	@echo '  make ephem-generate    - generate GTAB datasets (see scripts/ephem)'
	@echo '  make tidy              - go mod tidy for api'

all: test

test: api-test contracts-test

api-test:
	cd api && go test -race ./...

api-run:
	API_PORT?=8080
	cd api && API_PORT=$(API_PORT) EPHEM_MODE=mock go run ./cmd/server

api-run-file:
	API_PORT?=8080
	@if [ -z "$(EPHEM_TABLE_PATH)" ]; then echo 'EPHEM_TABLE_PATH required'; exit 1; fi
	cd api && API_PORT=$(API_PORT) EPHEM_MODE=file EPHEM_TABLE_PATH=$(EPHEM_TABLE_PATH) go run ./cmd/server

contracts-test:
	cd contracts && scarb test

gas-snapshot:
	cd contracts && scarb test > ../test_output.txt 2>&1 || (echo 'contract tests failed'; tail -n 100 ../test_output.txt; exit 1)
	python scripts/contracts/extract_gas.py --input test_output.txt --out docs/perf/gas_snapshot.json --fail-on-missing
	@echo 'Snapshot written to docs/perf/gas_snapshot.json'

gas-compare:
	python scripts/ci/compare_gas.py

frontend-dev:
	cd frontend && npm run dev

ephem-generate:
	python scripts/ephem/generate.py --out ephem

lint:
	@echo '(placeholder) add linters here'

tidy:
	cd api && go mod tidy
