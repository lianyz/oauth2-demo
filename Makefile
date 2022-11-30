.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	echo "building resource server..."
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./resource-server

.PHONY: run
run:
	./bin/amd64/resource-server

.PHONY: clean
clean:
	rm -rf ./bin
