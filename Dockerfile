FROM golang:1.17 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN make



FROM alpine:latest

RUN apk add --no-cache --upgrade ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/cmd/build/* ./

CMD ["./svc"]
