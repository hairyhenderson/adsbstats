FROM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o /adsbstats .

FROM alpine:latest
COPY --from=builder /adsbstats /usr/local/bin/adsbstats
ENTRYPOINT ["/usr/local/bin/adsbstats"]
