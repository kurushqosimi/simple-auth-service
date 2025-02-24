# Build stage
FROM golang:1.22.2-alpine as BUILDER

WORKDIR /app

RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/app/main.go


# Run stage
FROM alpine

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrate.linux-amd64 /app/migrate
COPY /config/config.yaml /app/config/config.yaml
# COPY app.env .

COPY wait-for.sh /app/wait-for.sh
COPY start.sh /app/start.sh

RUN chmod +x /app/wait-for.sh /app/start.sh

COPY migrations /app/migrations


RUN ls -la /app && ls -la /app/migrations

EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]