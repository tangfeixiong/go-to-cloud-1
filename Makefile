
LDFLAGS = "-X github.com/tangfeixiong/go-to-cloud-1/pkg/version.gitMajor=0 -X github.com/tangfeixiong/go-to-cloud-1/pkg/version.gitMinor=2"

GO_PATH = /work

OUT_DIR = _output

VERBOSE_FLAGS = $(GOFLAGS)

.PHONY: all install
all install:
	GOPATH=$(GO_PATH) go install $(VERBOSE_FLAGS) github.com/tangfeixiong/go-to-cloud-1/cmd/apaas

.PHONY: build
build:
	OUT_DIR := data/bin

	BUILD_FLAGS = -o $(OUT_DIR)/apaas -a $(VERBOSE_FLAGS)

	GOPATH=$(GO_PATH) go build $(BUILD_FLAGS) github.com/tangfeixiong/go-to-cloud-1/cmd/apaas

.PHONY: alpine-docker
alpine-docker:
	GOPATH=$(GO_PATH) CGO_ENABLED=0 go build -o build/docker/apaas --installsuffix cgo -a -v github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
	touch -m build/docker/apaas
	docker build -t tangfeixiong/gotopaas build/docker/

.PHONY: clean
clean: 
	rm $(OUT_DIR)/apaas
