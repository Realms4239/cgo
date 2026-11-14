GO ?= go
FRONTEND := web/frontend

.PHONY: build test vet figures frontend clean verify verify-fast verify-selftest hooks

build: frontend
	$(GO) build -o bin/cgo$(shell go env GOEXE) ./cmd/cgo

test:
	$(GO) vet ./...
	$(GO) test ./...
	cd $(FRONTEND) && bun run typecheck && bun run build && bunx vitest run && npx playwright test

vet:
	$(GO) vet ./...

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

# ---- eBPF maintainer targets (Linux 6.8 build host only) ----
BPF_DIR := probe/bpf
BPF_CLANG ?= clang-18
BPF_CFLAGS := -O2 -g -target bpf -D__TARGET_ARCH_x86 -Wall -Werror -I$(BPF_DIR)

.PHONY: bpf bpf-check

$(BPF_DIR)/vmlinux.h:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $@

bpf: $(BPF_DIR)/vmlinux.h
	$(BPF_CLANG) $(BPF_CFLAGS) -c $(BPF_DIR)/bpf_probe.c -o $(BPF_DIR)/bpf_probe.o
	sha256sum $(BPF_DIR)/bpf_probe.o | cut -d' ' -f1 > $(BPF_DIR)/bpf_probe.o.sha256

bpf-check:
	@cd $(BPF_DIR) && if [ ! -f bpf_probe.o ] || [ ! -f bpf_probe.o.sha256 ]; then \
		printf '%s\n' "bpf-check: SKIP — prébuilt absent (make bpf sur l'hôte Linux 6.8; CI: Task 3.2)"; \
		exit 2; \
	fi && test "$$(sha256sum bpf_probe.o | cut -d' ' -f1)" = "$$(cat bpf_probe.o.sha256)"
