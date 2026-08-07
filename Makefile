# staticcheck is pinned so that a new release cannot break CI on its own.
STATICCHECK_VERSION := 2025.1.1

BIN := $(CURDIR)/bin

.PHONY: all
all: fmtcheck vet staticcheck test

.PHONY: build
build:
	go build ./...

# gofmt -l lists the files needing formatting but exits 0 either way, so the
# output has to be turned into the failure.
.PHONY: fmtcheck
fmtcheck:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "these files need gofmt:"; \
		echo "$$out"; \
		exit 1; \
	fi

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck: $(BIN)/staticcheck
	$(BIN)/staticcheck ./...

$(BIN)/staticcheck:
	GOBIN=$(BIN) go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)

.PHONY: test
test:
	go test ./...

# Fuzz targets check the matcher against a brute force oracle. Both packages
# have one, and go test only fuzzes a single target at a time.
FUZZTIME ?= 60s

.PHONY: fuzz
fuzz:
	go test -run '^$$' -fuzz FuzzMatcher -fuzztime $(FUZZTIME) .
	go test -run '^$$' -fuzz FuzzMatcher -fuzztime $(FUZZTIME) ./acascii

.PHONY: bench
bench:
	go test -run '^$$' -bench . -benchmem ./...

.PHONY: clean
clean:
	rm -rf $(BIN)
