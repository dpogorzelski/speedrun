GITCOMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(shell git branch --show-current)

all: requirements linux darwin

clean:
	rm -rf dist

requirements:
	@go get

linux-amd64: requirements
	GOOS=linux GOARCH=amd64 go build -ldflags "-X speedrun/cmd.version=$(VERSION) -X speedrun/cmd.commit=$(GITCOMMIT) -X speedrun/cmd.date=$(DATE)" -o dist/linux/amd64/speedrun

linux-arm64: requirements
	GOOS=linux GOARCH=arm64 go build -ldflags "-X speedrun/cmd.version=$(VERSION) -X speedrun/cmd.commit=$(GITCOMMIT) -X speedrun/cmd.date=$(DATE)" -o dist/linux/arm64/speedrun

linux: requirements linux-amd64 linux-arm64

darwin-amd64: requirements
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X speedrun/cmd.version=$(VERSION) -X speedrun/cmd.commit=$(GITCOMMIT) -X speedrun/cmd.date=$(DATE)" -o dist/darwin/amd64/speedrun

darwin-arm64: requirements
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X speedrun/cmd.version=$(VERSION) -X speedrun/cmd.commit=$(GITCOMMIT) -X speedrun/cmd.date=$(DATE)" -o dist/darwin/arm64/speedrun

darwin: requirements darwin-amd64 darwin-arm64
