GOTOOLS = \
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v2
PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_TAGS?=tendermint
BUILD_FLAGS = -ldflags "-X github.com/bcbchain/tendermint/version.GitCommit=`git rev-parse --short=8 HEAD`"

all: dist

########################################
### Build

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/tendermint ./cmd/tendermint/

install:
	# Nothing to do

########################################
### Distribution

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' sh -c "'$(CURDIR)/scripts/dist.sh'"

########################################
### Download contract

# download contract tar file to chain pkgs
dc:
	@sh -c "'$(CURDIR)/scripts/download.sh'"

########################################
### Formatting, linting, and vetting

fmt:
	@go fmt ./...

.PHONY: all build install dist fmt