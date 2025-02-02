FROM golang:1.23.4

WORKDIR /app
COPY ./traderpool/go.mod ./traderpool/go.sum ./
RUN go mod tidy
COPY ./traderpool /app

RUN go build -o main .

CMD ["./main"]