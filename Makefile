install-tools:
	go install github.com/vektra/mockery/v2@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --go_out=internal/genproto --go-grpc_out=internal/genproto api/fibonacci.proto

mocks:
	go generate ./...

generate:
	make proto mocks

test:
	go test ./...

up:
	docker compose up -d --build

run-app:
	docker compose up -d --build app

down:
	docker compose down

lint:
	 staticcheck ./...