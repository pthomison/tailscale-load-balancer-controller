# syntax=docker/dockerfile:1.3-labs

# FROM tailscale/tailscale:stable

# RUN apk add curl bash

# RUN <<EOF
# curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
# install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
# EOF

# COPY <<COPYEOF /ip-monitor-entrypoint.sh
# #!/usr/bin/env bash

# sleep 10

# while true
# do
#     IP=\"\$(tailscale --socket /tmp/tailscaled.sock ip --4)\"

#     if [[ \"\$IP\" != \"\" ]];
#     then
#         echo "Update IP: \${IP}"
#         kubectl annotate pod \$HOSTNAME \"operator.pthomison.com/tailscale-ip=\${IP}\"
#     fi

#     sleep 30
# done
# COPYEOF

# RUN chmod +x /ip-monitor-entrypoint.sh

# ENTRYPOINT [ "/entrypoint.sh" ]


###########


# Build the manager binary
FROM golang:1.20 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
# COPY api/ api/
# COPY controllers/ controllers/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o ip-updater cmd/ip-watcher/ip-watcher.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM tailscale/tailscale:stable
RUN apk add curl bash

WORKDIR /
COPY --from=builder /workspace/ip-updater .
USER 65532:65532

