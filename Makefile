.PHONY: test
test:
	go vet ./...
	go test -race -short ./...

build:
	docker build . -t daxire/daxi-core/shortana