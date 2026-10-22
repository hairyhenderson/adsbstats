FROM golang:1.27.3-alpine3.24@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS builder
WORKDIR /src
COPY . .
RUN go build -o /adsbstats .

FROM alpine:3.24@sha256:c6b073f8e9e58d881f6ffba6ebf3b9ebd60b9e7d6e8337018c0e68bb5c356873
COPY --from=builder /adsbstats /usr/local/bin/adsbstats
ENTRYPOINT ["/usr/local/bin/adsbstats"]
