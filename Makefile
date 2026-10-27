GO ?= go
FRONTEND := web/frontend

.PHONY: build test vet test-real figures frontend clean verify verify-fast verify-selftest hooks

build: frontend
	$(GO) build -o bin/cgo$(shell go env GOEXE) ./cmd/cgo

test:
	$(GO) vet ./...
	$(GO) test ./...
	cd $(FRONTEND) && bun run typecheck && bun run build && bunx vitest run && npx playwright test

vet:
	$(GO) vet ./...

# Real gates (tc/netem/BBR) — run INSIDE the VM only.
test-real:
	$(GO) test -tags=real ./pkg/qdisc/... ./pkg/campagne/...

figures: build
	./bin/cgo figures

verify:
	python tools/verify.py

verify-fast:
	python tools/verify.py --fast

verify-selftest:
	python tools/verify.py --selftest

hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/pre-push 2>/dev/null || true
	python tools/verify.py --list

frontend:
	cd $(FRONTEND) && bun run build

clean:
	rm -rf bin $(FRONTEND)/dist
