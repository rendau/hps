FROM alpine:latest

RUN apk add --no-cache --upgrade ca-certificates tzdata curl

WORKDIR /app

COPY ./cmd/build/. ./

HEALTHCHECK --start-period=4s --interval=10s --timeout=2s --retries=3 CMD curl -f http://localhost:3003/healthcheck || false

CMD ["./svc"]
