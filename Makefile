.PHONY: build docker

build:
	cd signerService && GOOS=linux GOARCH=amd64 go build -o ../signingService && cd ../

	docker build -t signer-service .
