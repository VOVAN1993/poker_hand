FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o poker-hand ./cmd/poker-hand

EXPOSE 8080

CMD ["./poker-hand"]