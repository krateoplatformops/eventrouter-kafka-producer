# Build environment
# -----------------
FROM golang:1.19.2-bullseye as builder
LABEL stage=builder

ARG DEBIAN_FRONTEND=noninteractive

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y ca-certificates openssl git tzdata && apt-get install -y --no-install-recommends && \
  update-ca-certificates && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY internal/ internal/
COPY main.go main.go

# Build
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /bin/service main.go && \
    strip /bin/service

# Deployment environment
# ----------------------
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/service /service


ARG KAFKA_EVENTROUTER_CONSUMER_PORT

EXPOSE ${KAFKA_EVENTROUTER_CONSUMER_PORT}

ENTRYPOINT ["/service"]