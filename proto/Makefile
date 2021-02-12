all: client.pb.go

client.pb.go: client.proto
	protoc --go_out=plugins=grpc:. client.proto

.PHONY: clean
clean:
	rm client.pb.go
