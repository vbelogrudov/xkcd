build:
	go build -o xkcd ./cmd/xkcd

lint:
	golangci-lint run ./...
