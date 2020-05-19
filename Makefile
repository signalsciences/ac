all: ./bin/golangci-lint
	go build ./...
	./bin/golangci-lint run
	go test -cover ./...

./bin/golangci-lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v1.27.0
