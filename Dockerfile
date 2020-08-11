FROM golang:1.14.1 as build

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine

WORKDIR /app
COPY --from=build /app/main /app/main

ENTRYPOINT ["/app/main"]
CMD []