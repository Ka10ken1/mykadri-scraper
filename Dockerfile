FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o scraper ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/scraper .
COPY --from=builder /app/web ./web

EXPOSE 8080

ENV MONGO_URI=mongodb://mongo:27017
ENV MONGO_DB=mykadri
ENV MONGO_COLLECTION=movies

CMD ["./scraper"]

