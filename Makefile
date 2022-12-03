.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p bin/amd64

	@echo "building server..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./server

	@echo "building client..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 ./client

	@echo "copy static files..."
	mkdir -p bin/amd64/static
	cp ./server/static/*.html ./bin/amd64/static/
.PHONY: run
run: build
	clear
	./bin/amd64/server -d=false

.PHONY: run.client
run.client:
	clear
	curl "http://localhost:9096/register?clientId=CLIENT_12345&clientSecret=CLIENT_xxxxx&clientAddr=http://localhost:9094"
	./bin/amd64/client -id CLIENT_12345 -secret CLIENT_xxxxx -addr http://localhost:9094

.PHONY: clean
clean:
	rm -rf ./bin
