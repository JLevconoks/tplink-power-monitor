FROM golang:1.15-alpine as builder

ENV GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app .

FROM alpine
COPY --from=builder /build/app /

ENTRYPOINT ["/app"]