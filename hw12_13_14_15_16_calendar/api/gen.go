//go:build ignore

//go:generate protoc -I. -I./third_party/googleapis --go_out=gen --go-grpc_out=gen --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative EventService.proto
package api
