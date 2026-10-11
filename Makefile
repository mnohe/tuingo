.PHONY: test lint

test: ## unit tests
	@go test ./...

lint: ## gofmt and go vet, failing on any unformatted file
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt: files above need formatting"; exit 1; }
	@go vet ./...
