APP_NAME=kube-debugger

.PHONY: build test run

build:
	go build -o $(APP_NAME) ./main.go

test:
	go test ./...

run:
	go run ./main.go
