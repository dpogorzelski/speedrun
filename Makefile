GITCOMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(shell git branch --show-current)
DIST=dist
LINUX_PATH=$(DIST)/linux
DARWIN_PATH=$(DIST)/darwin

all: requirements speedrun-linux speedrun-darwin portal-linux

clean:
	rm -rf dist

requirements:
	@go mod tidy -compat=1.17
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

speedrun-linux-amd64: requirements
	GOOS=linux GOARCH=amd64 go build cmd/speedrun -ldflags "-X ./cmd/speedrun/cli.version=$(VERSION) -X ./cmd/speedrun/cli.commit=$(GITCOMMIT) -X ./cmd/speedrun/cli.date=-o $(DIST)/speedrun-linux-amd64 $(DATE)"

speedrun-linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build cmd/speedrun -ldflags "-X ./cmd/speedrun/cli.version=$(VERSION) -X ./cmd/speedrun/cli.commit=$(GITCOMMIT) -X ./cmd/speedrun/cli.date=-o $(DIST)/speedrun-linux-arm64 $(DATE)"

speedrun-linux: requirements speedrun-linux-amd64 speedrun-linux-arm64

speedrun-darwin-amd64: requirements
	GOOS=darwin GOARCH=amd64 go build cmd/speedrun -ldflags "-X ./cmd/speedrun/cli.version=$(VERSION) -X ./cmd/speedrun/cli.commit=$(GITCOMMIT) -X ./cmd/speedrun/cli.date=-o $(DIST)/speedrun-darwin-amd64 $(DATE)"

speedrun-darwin-arm64: requirements
	GOOS=darwin GOARCH=arm64 go build cmd/speedrun -ldflags "-X ./cmd/speedrun/cli.version=$(VERSION) -X ./cmd/speedrun/cli.commit=$(GITCOMMIT) -X ./cmd/speedrun/cli.date=-o $(DIST)/speedrun-darwin-arm64 $(DATE)"

speedrun-darwin: requirements speedrun-darwin-amd64 speedrun-darwin-arm64

speedrun: requirements speedrun-linux speedrun-darwin


portal-linux-amd64: requirements
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/portal-linux-amd64 ./cmd/portal
# zip $(DIST)/portal-linux-amd64.zip $(LINUX_PATH)/amd64/portal

portal-linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build -o $(DIST)/portal-linux-arm64 ./cmd/portal
# zip $(DIST)/portal-linux-arm64.zip $(LINUX_PATH)/arm64/portal

portal-linux: requirements portal-linux-amd64 portal-linux-arm64

portal: requirements portal-linux

proto:
	protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/portal/portal.proto