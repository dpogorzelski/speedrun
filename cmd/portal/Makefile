GITCOMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(shell git branch --show-current)
DIST=../dist
LINUX_PATH=$(DIST)/linux
DARWIN_PATH=$(DIST)/darwin

all: requirements linux

clean:
	rm -rf $(DIST)

requirements:
	@go get

linux-amd64: requirements
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_PATH)/amd64/portal
	zip $(DIST)/portal-linux-amd64.zip $(LINUX_PATH)/amd64/portal

linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build -o $(LINUX_PATH)/arm64/portal
	zip $(DIST)/portal-linux-arm64.zip $(LINUX_PATH)/arm64/portal

linux: requirements linux-amd64 linux-arm64

darwin-amd64: requirements
	GOOS=darwin GOARCH=amd64 go build -o $(DARWIN_PATH)/amd64/portal
	zip $(DIST)/portal-darwin-amd64.zip $(DARWIN_PATH)/amd64/portal

darwin-arm64: requirements
	GOOS=darwin GOARCH=arm64 go build -o $(DARWIN_PATH)/arm64/portal
	zip $(DIST)/portal-darwin-arm64.zip $(DARWIN_PATH)/arm64/portal

darwin: requirements darwin-amd64 darwin-arm64