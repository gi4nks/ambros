BINARY_NAME=ambros

build:
	GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME} cmd/main.go
	# GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux cmd/main.go
	# GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows cmd/main.go

run: build
	./bin/${BINARY_NAME}

clean:
	go clean
	rm ./bin/${BINARY_NAME}
	# rm ./bin/${BINARY_NAME}-linux
	# rm ./bin/${BINARY_NAME}-windows