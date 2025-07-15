all: clean lint check test-local-image

OS:=$(shell go env GOOS)

test: test-local-image
ifeq ($(OS),windows)
	@echo "Skipping test on windows, issue with -- and testscript"
else
	go test -race ./...
endif

lint: .golangci.yml
	golangci-lint run

tev:
	CGO_ENABLED=0 GOOS=linux go build -a .

clean:
	rm -f tev

.golangci.yml: Makefile
	curl -fsS -o .golangci.yml https://raw.githubusercontent.com/fortio/workflows/main/golangci.yml


.PHONY: lint check all clean
