.PHONY: init
# init dependcy
init:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1

.PHONY: build
# build protoc-gen-go-errors
build:
	go build -o protoc-gen-go-errors *.go && mv protoc-gen-go-errors ${GOPATH}/bin/

.PHONY: generate
# generate error test code
generate:
	@cd ./gerr && protoc -I . \
		--go_out=paths=source_relative:. \
		errors.proto

.PHONY: test
# generate error test code
test:
	@cd ./test && protoc -I . -I ../gerr \
		--go_out=paths=source_relative:. \
        --go-errors_out=paths=source_relative:. \
		test.proto && \
	go test ./...