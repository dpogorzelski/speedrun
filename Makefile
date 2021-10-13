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
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/speedrun-linux-amd64 -ldflags "-X github.com/speedrunsh/speedrun/cmd/speedrun/cli.version=$(VERSION) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.commit=$(GITCOMMIT) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.date=$(DATE)" ./cmd/speedrun

speedrun-linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build -o $(DIST)/speedrun-linux-arm64 -ldflags "-X github.com/speedrunsh/speedrun/cmd/speedrun/cli.version=$(VERSION) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.commit=$(GITCOMMIT) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.date=$(DATE)" ./cmd/speedrun

speedrun-linux: speedrun-linux-amd64 speedrun-linux-arm64

speedrun-darwin-amd64: requirements
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/speedrun-darwin-amd64 -ldflags "-X github.com/speedrunsh/speedrun/cmd/speedrun/cli.version=$(VERSION) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.commit=$(GITCOMMIT) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.date=$(DATE)" ./cmd/speedrun

speedrun-darwin-arm64: requirements
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/speedrun-darwin-arm64 -ldflags "-X github.com/speedrunsh/speedrun/cmd/speedrun/cli.version=$(VERSION) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.commit=$(GITCOMMIT) -X github.com/speedrunsh/speedrun/cmd/speedrun/cli.date=$(DATE)" ./cmd/speedrun

speedrun-darwin: speedrun-darwin-amd64 speedrun-darwin-arm64

speedrun: speedrun-linux speedrun-darwin


portal-linux-amd64: requirements
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/portal-linux-amd64 ./cmd/portal
#	zip $(DIST)/portal-linux-amd64.zip $(DIST)/portal-linux-amd64

portal-linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build -o $(DIST)/portal-linux-arm64 ./cmd/portal
#	zip $(DIST)/portal-linux-arm64.zip $(DIST)/portal-linux-arm64

portal-linux: portal-linux-amd64 portal-linux-arm64

portal: portal-linux

dev: speedrun-darwin-amd64 portal-linux-amd64

proto:
	protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/portal/portal.proto