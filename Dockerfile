FROM golang:1.19 as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main -v cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/app/main .
# COPY main .
COPY config.yaml .
COPY templates/ templates/

CMD ["./main"]
