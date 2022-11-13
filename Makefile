.PHONY: build run clean build-and-run build-docker run-docker build-and-run-docker tidy test

BINARY_NAME=THEDUTCHAPP

build:
	go build -o $(BINARY_NAME) cmd/server/main.go

run:
	./$(ARGS)

run-no-build:
	go run cmd/server/main.go

clean:
	if [ -f $(BINARY_NAME) ] ; then rm $(BINARY_NAME) ; fi

build-and-run: build run

build-docker:
	docker build -t $(BINARY_NAME) .

run-docker:
	docker run -p 8080:8080 $(BINARY_NAME)

build-and-run-docker: build-docker run-docker

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

get:
	go get -u $(PKG)

tidy:
	go mod tidy

dep:
	go mod download

lint:
	go fmt ./...

vet:
	go vet ./...