.PHONY:help
help: ## Print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

gen-proto: ## generate golang implementation
	@type protoc-gen-go > /dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go
	protoc --go_out=. proto_test.proto