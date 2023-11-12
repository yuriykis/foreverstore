BINARY_NAME=fs

build:
	@go build -o bin/${BINARY_NAME} -v

run: build
	@./bin/${BINARY_NAME}

test:
	@go test -v ./...