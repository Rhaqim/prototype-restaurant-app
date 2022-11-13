FROM golang:1.19

WORKDIR /go/src/app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o THEDUTCHAPP cmd/server/main.go

EXPOSE 8080

CMD ["./THEDUTCHAPP"]