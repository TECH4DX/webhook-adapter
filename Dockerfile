FROM golang:alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o server src/server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /app ./
CMD ["./server"]