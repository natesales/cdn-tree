DIST_DIR := dist
VERSION := $(shell date +%FT%T%z)-$(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := '-X main.version=$(VERSION)'

all: clean proto api client

clean:
	rm -rf $(DIST_DIR)

proto:
	cd proto && make

api:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/api cmd/api/api.go

client:
	go build -ldflags $(LDFLAGS) -o $(DIST_DIR)/client cmd/client/client.go

controller-install:
	cd cbootstrap && ansible-playbook -i hosts.yml install.yml

controller-deploy:
	cd cbootstrap && ansible-playbook -i hosts.yml deploy.yml
