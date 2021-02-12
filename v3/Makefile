DIST_DIR=dist

all: clean build-api build-client

clean:
	rm -rf $(DIST_DIR)

build-api:
	go build -o $(DIST_DIR)/api cmd/api/api.go

build-client:
	go build -o $(DIST_DIR)/client cmd/client/client.go
