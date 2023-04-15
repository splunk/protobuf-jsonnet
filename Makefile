PROTOC_GEN_VALIDATE_VERSION=v0.9.1

local:
	make build
	make test

build:
	go install ./...

test:
	go test -v ./...

.bin/protoc-gen-go:
	mkdir -p .bin/
	GOBIN=$(PWD)/.bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1

generate: .bin/protoc-gen-go
	rm -rf .gen
	git clone --depth 1 --branch $(PROTOC_GEN_VALIDATE_VERSION) git@github.com:envoyproxy/protoc-gen-validate.git .gen
	@PATH=$(PWD)/.bin:$(PATH) protoc --go_out=. --go_opt=Mvalidate/validate.proto=internal/validate  -I .gen/ validate/validate.proto
	mv .gen/validate/validate.proto internal/validate/
	rm -rf .gen
