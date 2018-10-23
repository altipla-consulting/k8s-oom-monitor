
FROM golang:1.11 as builder

WORKDIR /k8s-oom-monitor
COPY . .

RUN go install .

# ==============================================================================

FROM launcher.gcr.io/google/debian9:latest

RUN apt-get update && \
    apt-get install -y ca-certificates

COPY --from=builder /go/bin/k8s-oom-monitor /opt/k8s-oom-monitor

ENTRYPOINT ["/opt/k8s-oom-monitor"]
