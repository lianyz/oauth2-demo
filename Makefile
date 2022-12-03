.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p bin/amd64

	@echo "building server..."
	CGO_ENABLED=0 GOARCH=amd64 go build -o bin/amd64 ./server

	@echo "building client..."
	CGO_ENABLED=0 GOARCH=amd64 go build -o bin/amd64 ./client

	@echo "copy static files..."
	mkdir -p bin/amd64/static
	cp ./server/static/*.html ./bin/amd64/static/

SERVER_ADDR = 127.0.0.1
CLIENT_ADDR = 127.0.0.1

.PHONY: run
run: build
	clear
	./bin/amd64/server -d=false -ip=$(SERVER_ADDR)

.PHONY: run.client
run.client:
	clear
	curl "http://$(SERVER_ADDR):9096/register?clientId=CLIENT_12345&clientSecret=CLIENT_xxxxx&clientAddr=http://$(CLIENT_ADDR):9094"
	./bin/amd64/client -id CLIENT_12345 -secret CLIENT_xxxxx -addr http://$(CLIENT_ADDR):9094 -server http://$(SERVER_ADDR):9096

.PHONY: install.webhook
install.webhook:
	@echo "add token to webhook-config/config under oauth2-user"
	cp ~/.kube/config ~/.kube/config.bak
	cp ./webhook-config/config ~/.kube/

	cp /etc/kubernetes/manifests/kube-apiserver.yaml /etc/kubernetes/manifests/kube-apiserver.yaml.bak
	cp ./webhook-config/kube-apiserver.yaml /etc/kubernetes/manifests/

	cp /etc/config/webhook-config.json /etc/config/webhook-config.json.bak
	cp ./webhook-config/webhook-config.json /etc/config/
.PHONY: clean
clean:
	rm -rf ./bin
