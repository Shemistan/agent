lsof:
	lsof -i :8080

kill:
	kill -9 20906

lint:
	golangci-lint cache clean
	golangci-lint run ./...
