BINARY_NAME=app
MONGO_CONTAINER_NAME=biocad
host = localhost
port = 7999
sourceDir = files
outputDir = processed
dbhost = 127.0.0.1
dbport = 7998

.PHONY:
	build \
	run \
	dropdb \

build:
	go build -o $(BINARY_NAME) cmd/api-server/*.go

run: build
	sudo docker run -d -p $(dbhost):$(dbport):27017 --name=$(MONGO_CONTAINER_NAME)  mongodb/mongodb-community-server
	./$(BINARY_NAME) -host=$(host) -port=$(port) -sourceDir=$(sourceDir) -outputDir=$(outputDir) -dbhost=$(dbhost) -dbport=$(dbport)

dropdb:
	sudo docker stop biocad && sudo docker rm biocad
