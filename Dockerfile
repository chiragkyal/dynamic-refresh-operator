FROM golang:1.21 as builder
WORKDIR /go/src/github.com/chiragkyal/dynamic-refresh-operator
COPY . .
RUN make GO_REQUIRED_MIN_VERSION:=

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3-1612
COPY --from=builder /go/src/github.com/chiragkyal/dynamic-refresh-operator/dynamic-refresh-operator /usr/bin/
ENTRYPOINT ["/usr/bin/dynamic-refresh-operator"]
LABEL io.k8s.display-name="OpenShift Dynamic Refresh Operator" \
    io.k8s.description="The Dynamic Refresh Operator installs and maintains the Dynamic Refresh on a cluster."
