# Build the ipvs-node-controller binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build ipvs-node-controller
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o ipvs-node-controller main.go

# Build image
FROM alpine:3.11.6
RUN apk add iptables=1.8.3-r2

WORKDIR /
COPY --from=builder /workspace/ipvs-node-controller .
ENTRYPOINT ["/ipvs-node-controller"]
