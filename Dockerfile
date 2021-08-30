# Build the network-node-manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build network-node-manager
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o network-node-manager main.go

# Build image
FROM alpine:3.13.3
RUN apk add --no-cache iptables=1.8.6-r0 ip6tables=1.8.6-r0
COPY scripts/iptables-wrapper-installer.sh /
RUN chmod 0744 /iptables-wrapper-installer.sh && /iptables-wrapper-installer.sh

WORKDIR /
COPY --from=builder /workspace/network-node-manager .
ENTRYPOINT ["/network-node-manager"]
