FROM golang AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o terminal ./cmd/api/main.go

FROM scratch

COPY --from=builder /app/terminal .

EXPOSE 8080 

CMD ["./terminal"]