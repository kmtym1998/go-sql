OUTPUT := ${GOPATH}/bin

.PHONY: build
build: ## バイナリをビルド
	go build -o ${OUTPUT}


.PHONY: fmt
fmt: ## フォーマット
	go fmt ./...
