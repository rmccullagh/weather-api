.PHONY: run docs coverage all

coverage:
	go test $(shell go list ./... | grep -v /docs/) -coverpkg=./... -coverprofile ./coverage.out
	go tool cover -func ./coverage.out
run:
	go run main.go
docs:
	~/go/bin/swag init
all:
	make run
	make docs
