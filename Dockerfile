FROM golang:1.26-alpine3.24@sha256:3082400e369fa24d5fc60bca20edab3f6d604e0c5a690ec66b295eff4dd87ade AS builder
WORKDIR /src
COPY . .
RUN go build -o /adsbstats .

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
COPY --from=builder /adsbstats /usr/local/bin/adsbstats
ENTRYPOINT ["/usr/local/bin/adsbstats"]
