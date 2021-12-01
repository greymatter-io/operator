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

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator main.go

# Add CLI
FROM docker.greymatter.io/release/greymatter:3.0.0 as cli
# FROM docker.greymatter.io/internal/cli:4.0.0-preview as cli

# Use distroless as minimal base image to package the operator and greymatter CLI binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/operator .
COPY --from=builder /workspace/pkg/version/cue.mod/ cue.mod/
COPY --from=cli /bin/greymatter /bin/greymatter
# COPY --from=cli /usr/local/bin/greymatter /bin/greymatter
USER 65532:65532

ENTRYPOINT ["/operator"]
