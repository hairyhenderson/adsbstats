FROM golang:1.26-alpine3.24@sha256:ce864e7223ac17b1775e6fd0b4c0db580c2eb50e7953a427916379e4b92a1628 AS builder
WORKDIR /src
COPY . .
RUN go build -o /adsbstats .

FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
COPY --from=builder /adsbstats /usr/local/bin/adsbstats
ENTRYPOINT ["/usr/local/bin/adsbstats"]
