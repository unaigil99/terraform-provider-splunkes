default: build

build:
	go install

test:
	go test ./... -timeout 30s -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w .

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	golangci-lint run ./...

vet:
	@echo "go vet ."
	@go vet ./... ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs."; \
		exit 1; \
	fi

docs:
	go generate ./...

.PHONY: build test testacc fmt fmtcheck lint vet docs
