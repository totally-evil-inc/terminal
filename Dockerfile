FROM golang:1.25-alpine AS air

WORKDIR /app

RUN go install github.com/air-verse/air@latest

FROM golang:1.25-alpine AS dev

WORKDIR /app

COPY --from=air /go/bin/air /usr/local/bin/air

COPY go.mod ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o terminal ./cmd/api/main.go

FROM scratch

COPY --from=builder /app/terminal .

EXPOSE 8080 

CMD ["./terminal"]
