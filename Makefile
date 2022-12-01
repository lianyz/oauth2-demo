.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	@echo "building resource server..."
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./resource-server

	@echo "building server..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./server

	@echo "building client..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./client

.PHONY: run
run:
	./bin/amd64/server

.PHONY: run.client
run.client:
	./bin/amd64/client

.PHONY: clean
clean:
	rm -rf ./bin
