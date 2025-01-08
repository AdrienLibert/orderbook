FROM golang:1.23.4

WORKDIR /app
COPY ./orderbook/go.mod ./orderbook/go.sum ./
RUN go mod tidy
COPY ./orderbook /app

RUN go build -o main .

CMD ["./main"]