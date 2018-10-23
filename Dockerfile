
FROM golang:1.11 as builder

WORKDIR /k8s-oom-monitor
COPY . .

RUN go install .

# ==============================================================================

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/k8s-oom-monitor /go/bin/k8s-oom-monitor

ENTRYPOINT ["/go/bin/k8s-oom-monitor"]
