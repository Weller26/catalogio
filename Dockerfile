FROM golang:1.27 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /app/catalogio ./cmd/catalogio/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/catalogio .
COPY --from=build /app/web ./web
EXPOSE 8080

CMD ["./catalogio"]