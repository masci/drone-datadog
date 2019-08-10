FROM golang:1.12 as builder

ARG SHA1
ARG TAG

ENV SHA1=$SHA1
ENV TAG=$TAG
ENV CGO_ENABLED 0

ADD . /go

RUN go build -o drone-datadog \
    -ldflags "-s -w -extldflags \"-static\" -X main.BuildCommit=$SHA1 -X main.BuildTag=$TAG" .
RUN ./drone-datadog -v

FROM plugins/base:linux-amd64
COPY --from=builder /go/drone-datadog /bin/drone-datadog
ENTRYPOINT ["/bin/drone-datadog"]
