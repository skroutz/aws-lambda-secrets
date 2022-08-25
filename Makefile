GOOS := "linux"

mod:
	@go mod tidy

fetch-secrets-amd64:
	GOOS=$(GOOS) GOARCH=amd64 go build \
		-o extension/fetch-secrets \
		cmd/fetch-secrets/main.go

load-secrets-amd64:
	GOOS=$(GOOS) GOARCH=amd64 go build \
		-o extension/wrapper/load-secrets \
		cmd/load-secrets/main.go

build-amd64: clean mod fetch-secrets-amd64 load-secrets-amd64

zip-amd64: build-amd64
	zip -r aws-lambda-secrets-amd64.zip extension/
	@echo "Extension amd64 zip archive created"


fetch-secrets-arm64:
	GOOS=$(GOOS) GOARCH=arm64 go build \
		-o extension/fetch-secrets \
		cmd/fetch-secrets/main.go

load-secrets-arm64:
	GOOS=$(GOOS) GOARCH=arm64 go build \
		-o extension/wrapper/load-secrets \
		cmd/load-secrets/main.go

build-arm64: clean mod fetch-secrets-arm64 load-secrets-arm64

zip-arm64: build-arm64
	zip -r aws-lambda-secrets-arm64.zip extension/
	@echo "Extension arm64 zip archive created"

release: zip-amd64 zip-arm64

clean:
	-rm -rf extension

.PHONY: build zip clean mod