.PHONY: build-web build test dev-web dev-server run

build-web:
	cd web && npm install && npm run build
	rm -rf cmd/server/web
	cp -r web/dist cmd/server/web

build: build-web
	go build -o bin/server ./cmd/server

test:
	go test ./...

dev-web:
	cd web && npm run dev

dev-server:
	go run ./cmd/server

run: build
	./bin/server
		telvie-application-service/
       telvie-control-service/
       telvie-iam-service/
       telvie-ingestion-service/
       telvie-metadata-service/
       telvie-notification-service/

