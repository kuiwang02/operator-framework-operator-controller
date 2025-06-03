all: update build
.PHONY: all
OUT_DIR=bin

GO_PKG_NAME := github.com/openshift-eng/openshift-tests-extension

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE_STATE := $(shell if git diff --quiet; then echo clean; else echo dirty; fi)

GO_LD_FLAGS := -s -w \
			-X '$(GO_PKG_NAME)/pkg/version.CommitFromGit=$(GIT_COMMIT)' \
			-X '$(GO_PKG_NAME)/pkg/version.BuildDate=$(BUILD_DATE)' \
			-X '$(GO_PKG_NAME)/pkg/version.GitTreeState=$(GIT_TREE_STATE)'

ifeq ($(shell go env GOOS),darwin)
	GO_LD_FLAGS += -extldflags='-ld_classic'
endif

build:
	mkdir -p "${OUT_DIR}"
	export GO111MODULE="on" && export GOFLAGS="" && export GOWORK=off && \
	{ GO_COMPLIANCE_POLICY="exempt_all" go build  -ldflags="${GO_LD_FLAGS}" -mod=vendor -o "${OUT_DIR}" "./cmd/olmv1-tests-ext";sed -i'' -e '/^toolchain go/d' go.mod; rm -f go.mod-e; }
.PHONY: build

go-mod-tidy:
	./hack/go-mod-tidy.sh
.PHONY: go-mod-tidy

# Include the library makefile
include $(addprefix ./, bindata.mk)


# This will call a macro called "add-bindata" which will generate bindata specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - input dirs
# $3 - prefix
# $4 - pkg
# $5 - output
# It will generate targets {update,verify}-bindata-$(1) logically grouping them in unsuffixed versions of these targets
# and also hooked into {update,verify}-generated for broader integration.
$(call add-bindata,testdata,test/extended/testdata/...,testextended,testdata,./test/extended/testdata/bindata.go)


clean:
	$(RM) ./bin/olmv1-tests-ext
.PHONY: clean
