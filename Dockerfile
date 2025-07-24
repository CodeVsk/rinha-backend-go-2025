FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

WORKDIR /app/cmd/api
RUN go build -o /rinha-app

FROM alpine:latest
COPY --from=builder /rinha-app /rinha-app

EXPOSE 8080

CMD ["/rinha-app"]