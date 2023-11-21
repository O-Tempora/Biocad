BINARY_NAME=app

.PHONY:
	build \
	run \

build:
	go build -o $(BINARY_NAME) cmd/api-server/*.go

run: build
	./$(BINARY_NAME)