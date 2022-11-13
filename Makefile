.PHONY: build run clean build-and-run build-docker run-docker build-and-run-docker tidy test

BINARY_NAME=THEDUTCHAPP

build:
	go build -o $(BINARY_NAME) cmd/server/main.go

run:
	./$(ARGS)

clean:
	if [ -f $(BINARY_NAME) ] ; then rm $(BINARY_NAME) ; fi

build-and-run: build run

build-docker:
	docker build -t $(BINARY_NAME) .

run-docker:
	docker run -p 8080:8080 $(BINARY_NAME)

build-and-run-docker: build-docker run-docker

tidy:
	go mod tidy

test:
	go test -v ./...

get:
	go get -u $(PKG)
