FROM golang AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

FROM golang AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o terminal ./cmd/api/main.go

FROM scratch

COPY --from=builder /app/terminal .

EXPOSE 8080 

CMD ["./terminal"]
