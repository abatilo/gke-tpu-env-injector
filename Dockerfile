FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download -x

COPY main.go ./
# Purposefully set AFTER downloading and caching dependencies
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /go/bin/gke-tpu-env-injector .

FROM scratch
COPY --from=builder /go/bin/gke-tpu-env-injector /usr/local/bin/gke-tpu-env-injector

EXPOSE 8000
ENTRYPOINT ["gke-tpu-env-injector"]
