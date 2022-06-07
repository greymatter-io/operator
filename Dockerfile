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

# Ensure an SSH key is available and trusts GitHub
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts && \
    chmod 444 /root/.ssh/known_hosts

FROM ubuntu:21.04

WORKDIR /app
COPY --from=builder /root/.ssh/known_hosts /app/known_hosts
COPY --from=builder /workspace/operator /app/operator
COPY --from=builder /workspace/greymatter /bin/greymatter
COPY --from=builder /workspace/pkg/cuemodule/core /app/core
RUN mkdir /app/fetched_cue && chmod 777 /app/fetched_cue && chown 1000:1000 /app/known_hosts
USER 1000:1000
ENV HOME=/app
ENV SSH_KNOWN_HOSTS=/app/known_hosts

CMD ["/app/operator"]
