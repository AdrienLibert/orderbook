FROM golang:1.23.4

WORKDIR /app
COPY ./orderbook_v2/go.mod ./orderbook_v2/go.sum ./
RUN go mod tidy
COPY ./orderbook_v2 /app

RUN go build -o main .

CMD ["./main"]