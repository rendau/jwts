.DEFAULT_GOAL := build

BINARY_NAME = svc
BUILD_PATH = cmd/build

build:
	mkdir -p $(BUILD_PATH)
	CGO_ENABLED=0 go build -o $(BUILD_PATH)/$(BINARY_NAME) cmd/main.go

clean:
	rm -rf $(BUILD_PATH)

generate-proto-jwts_v1:
	mkdir -p pkg/proto
	protoc -I vendor-proto -I api/proto \
	--go_out pkg/proto --go_opt paths=source_relative \
		--go_opt=Mcommon/common.proto=`go list -m`/pkg/proto/common \
	--go-grpc_out pkg/proto --go-grpc_opt paths=source_relative \
	--grpc-gateway_out pkg/proto --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=json_names_for_fields=false:docs \
	api/proto/jwts_v1/*.proto

generate-proto: generate-proto-jwts_v1
