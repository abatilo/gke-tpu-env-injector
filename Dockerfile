FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x

COPY config.go ./
COPY main.go ./
# Purposefully set AFTER downloading and caching dependencies
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /go/bin/gke-tpu-env-injector .

FROM scratch
COPY --from=builder /go/bin/gke-tpu-env-injector /usr/local/bin/gke-tpu-env-injector

EXPOSE 443
ENTRYPOINT ["gke-tpu-env-injector"]
