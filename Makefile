
LDFLAGS :=

OUT_DIR = _output

QY_GOFLAGS = $(GOFLAGS)

.PHONY: all install
all install:
	go install $(QY_GOFLAGS) github.com/tangfeixiong/go-to-cloud-1/cmd/ociacibuilds

.PHONY: build
build:
	go build $(QY_GOFLAGS) -o $(OUT_DIR)/ociacibuilds github.com/tangfeixiong/go-to-cloud-1/cmd/ociacibuilds

.PHONY: clean
clean: 
	rm $(OUT_DIR)/ociacibuilds
