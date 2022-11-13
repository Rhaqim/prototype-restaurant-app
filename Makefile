run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

compile:
	echo "Compiling for linux"
	GOOS=linux GOARCH=amd64 go build -o bin/server cmd/server/main.go
