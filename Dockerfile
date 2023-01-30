FROM golang:1.19

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . /app

# RUN go build -o THEDUTCHAPP cmd/server/main.go

EXPOSE 8080

# CMD ["./THEDUTCHAPP"]

# CMD ["go", "run", "cmd/server/main.go"]