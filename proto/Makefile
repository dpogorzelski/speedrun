.PHONY: requirements portal

all: requirements portal

requirements:
	@go mod download
	@go install storj.io/drpc/cmd/protoc-gen-go-drpc@latest

portal:
	protoc --go_out=paths=source_relative:. --go-drpc_out=paths=source_relative:. portal/*.proto
