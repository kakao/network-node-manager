# Build the network-node-manager binary
FROM golang:1.16.7 as builder

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
FROM alpine:3.14.2
RUN apk add --no-cache iptables=1.8.7-r1 ip6tables=1.8.7-r1
COPY scripts/iptables-wrapper-installer.sh /
RUN chmod 0744 /iptables-wrapper-installer.sh 
RUN /iptables-wrapper-installer.sh

WORKDIR /
COPY --from=builder /workspace/network-node-manager .
ENTRYPOINT ["/network-node-manager"]
