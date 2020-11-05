default: build

SRCS    := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-w -s"

test:
	go test -v -race ./...

build: $(SRCS)
	CGO_ENABLED=0 go build -o fountain -trimpath $(LDFLAGS)

clean:
	rm -rf fountain
