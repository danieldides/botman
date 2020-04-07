FROM golang:1.12 AS builder
WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . /app
RUN go build -o botman

FROM gcr.io/distroless/base-debian10:debug
WORKDIR /app
COPY --from=builder /app/botman /app/botman
ENTRYPOINT ["./botman"]