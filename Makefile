@PHONY: dev
dev:
	CONF_FILE_PATH=./config/dev.yaml go run main.go

@PHONY: test
test:
	go test ./...

@PHONY: test-e2e
test-e2e:
	go test -v -timeout 30s ./e2e/
