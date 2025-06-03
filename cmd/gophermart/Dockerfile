FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/air-verse/air@latest

COPY ../../ ./

RUN mkdir -p /app/logs

CMD ["air", "-c", ".air.server.toml"]
