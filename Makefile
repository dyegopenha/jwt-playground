.PHONY: run
run:
	@go run cmd/server/main.go

.PHONY: generate
generate:
	@go generate ./...
