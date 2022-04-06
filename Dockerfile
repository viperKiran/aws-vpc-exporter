FROM golang:1.17.6-alpine@sha256:519c827ec22e5cf7417c9ff063ec840a446cdd30681700a16cf42eb43823e27c AS build

# renovate: datasource=go depName=github/prometheus/promu
ARG PROMU_VERSION=v0.13.0

WORKDIR /go/src/aws-vpc-exporter

RUN apk add --no-cache git \
    && go install "github.com/prometheus/promu@${PROMU_VERSION}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN promu build --verbose

FROM quay.io/prometheus/busybox:latest@sha256:2548dd93c438f7cf8b68dc2ff140189d9bcdae7130d3941524becc31573ec9e3

COPY --from=build /go/src/aws-vpc-exporter/aws-vpc-exporter /bin/aws-vpc-exporter

USER 1000:1000

ENTRYPOINT ["/bin/aws-vpc-exporter"]
