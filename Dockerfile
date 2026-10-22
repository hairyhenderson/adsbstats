FROM golang:1.27.3-alpine3.24@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS builder
WORKDIR /src
COPY . .
RUN go build -o /adsbstats .

FROM alpine:3.24@sha256:71e9b7bae160ba167fa17414401b94cbcdaf3c32feeab9298bd794a2a0614c35
COPY --from=builder /adsbstats /usr/local/bin/adsbstats
ENTRYPOINT ["/usr/local/bin/adsbstats"]
