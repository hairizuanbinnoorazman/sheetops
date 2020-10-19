# Build the manager binary
FROM golang:1.14 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY logger logger/
COPY appmgr appmgr/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager .

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM debian:buster-20200224
RUN apt update && apt install -y ca-certificates
WORKDIR /
COPY --from=builder /workspace/manager .
ADD sheetops-auth.json /sheetops-auth.json
# USER nonroot:nonroot

ENTRYPOINT ["/manager"]
