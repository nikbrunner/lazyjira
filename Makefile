.PHONY: build build-version build-demo lint lint-fix lint-docs vet test clean check startup-smoke check-demo tidy fix preview e2e e2e-gen e2e-update hooks check-staged

build:
	go build -o lazyjira ./cmd/lazyjira

hooks:
	mise install lefthook
	mise exec lefthook -- lefthook install

build-demo:
	go build -tags demo -o lazyjira ./cmd/lazyjira

build-version:
	go build -ldflags "-s -w -X main.version=$$(git rev-parse --short HEAD)" -o lazyjira ./cmd/lazyjira

lint:
	go tool golangci-lint run ./...

lint-fix:
	go tool golangci-lint fmt
	go tool golangci-lint run --fix ./...

lint-docs:
	npx --yes markdownlint-cli README.md CHANGELOG.md docs/*.md --disable MD001 MD013 MD024 MD033 MD040 MD041 MD060

vet:
	go vet ./...

test:
	go test -race ./...

clean:
	rm -f lazyjira

check: lint vet build test startup-smoke

startup-smoke: build
	python3 e2e/startup_smoke.py

check-staged:
	@bash -euo pipefail -c '\
		status=0; \
		while IFS= read -r -d "" path; do \
			if ! unformatted=$$(git show ":$$path" | gofmt -l); then \
				printf "Cannot check staged Go file: %s\n" "$$path" >&2; \
				status=1; \
			elif [[ -n "$$unformatted" ]]; then \
				printf "Staged Go file needs gofmt: %s\n" "$$path" >&2; \
				status=1; \
			fi; \
		done < <(git diff --cached --name-only --diff-filter=ACMR -z -- "*.go"); \
		exit "$$status"'

tidy:
	go mod tidy

fix: tidy lint-fix

check-demo:
	go tool golangci-lint run --build-tags demo ./...
	go vet -tags demo ./...
	go build -tags demo -o lazyjira ./cmd/lazyjira

preview: build-demo e2e-gen
	@vhs -q e2e/tapes/00_preview.tape
	@rm -f e2e/tapes/*.tape

e2e: build-demo e2e-gen
	@pids=""; fail=0; \
	for tape in e2e/tapes/*.tape; do \
		echo "Running $$tape..."; \
		vhs -q $$tape & pids="$$pids $$!"; \
	done; \
	for pid in $$pids; do \
		wait $$pid || fail=1; \
	done; \
	if [ $$fail -eq 1 ]; then echo "SOME TAPES FAILED" && exit 1; fi
	@rm -f e2e/tapes/*.tape e2e/tapes/*.gif
	@echo "All tapes passed."

e2e-gen:
	@./e2e/tape.sh generate-all

e2e-update: build-demo e2e-gen
	@pids=""; \
	for tape in e2e/tapes/*.tape; do \
		echo "Running $$tape..."; \
		vhs -q $$tape & pids="$$pids $$!"; \
	done; \
	for pid in $$pids; do wait $$pid; done
	@rm -f e2e/tapes/*.tape e2e/tapes/*.gif
	@echo "Recordings updated. Review with: git diff docs/assets/recordings/"
