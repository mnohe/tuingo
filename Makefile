.PHONY: test lint docs

test: ## unit tests
	@go test ./...

lint: ## gofmt and go vet, failing on any unformatted file
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt: files above need formatting"; exit 1; }
	@go vet ./...

docs: ## redraw the pictures in docs/ from the primitives' Draw methods
	@go run ./internal/docgen docs
