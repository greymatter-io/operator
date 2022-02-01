# Build the operator binary
FROM docker.io/golang:1.17 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Ensure CLI is downloaded to workspace
ARG username
ARG password
ENV USERNAME=$username
ENV PASSWORD=$password
COPY scripts/cli cli
RUN ./cli

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator main.go

FROM ubuntu:21.04

WORKDIR /app
COPY --from=builder /workspace/operator /app/operator
COPY --from=builder /workspace/pkg/cuemodule/cue.mod /app/cue.mod
COPY --from=builder /workspace/pkg/cuemodule/base /app/base
COPY --from=builder /workspace/pkg/cuemodule/meshconfigs /app/meshconfigs
COPY --from=builder /workspace/pkg/cuemodule/greymatter-cue /app/greymatter-cue
COPY --from=builder /workspace/greymatter /bin/greymatter
USER 1000:1000

ENTRYPOINT ["/app/operator"]
