GO ?= go
FRONTEND := web/frontend

.PHONY: build test vet test-real figures frontend clean

build: frontend
	$(GO) build -o bin/cgo$(shell go env GOEXE) ./cmd/cgo

test:
	$(GO) test ./...
	cd $(FRONTEND) && bunx vitest run && npx playwright test

vet:
	$(GO) vet ./...

# Real gates (tc/netem/BBR) — run INSIDE the VM only.
test-real:
	$(GO) test -tags=real ./pkg/qdisc/... ./pkg/campagne/...

figures: build
	./bin/cgo figures

frontend:
	cd $(FRONTEND) && bun run build

clean:
	rm -rf bin $(FRONTEND)/dist
