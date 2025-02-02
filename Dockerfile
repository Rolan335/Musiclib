FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o musiclib ./cmd/main.go

EXPOSE 8080

COPY .env.docker .env

ENV $(cat .env | xargs)

CMD ["./musiclib"]