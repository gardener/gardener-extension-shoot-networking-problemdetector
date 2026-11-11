# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.27.4@sha256:08bcb99a38f85a265f9a93d1d3323e2440eec333380f3ea26ee0dec9264bb41c AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-shoot-networking-problemdetector

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION
RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# gardener-extension-shoot-networking-problemdetector
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329 AS gardener-extension-shoot-networking-problemdetector

WORKDIR /
COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-shoot-networking-problemdetector /gardener-extension-shoot-networking-problemdetector
ENTRYPOINT ["/gardener-extension-shoot-networking-problemdetector"]

############# gardener-extension-shoot-networking-problemdetector-admission
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329 AS gardener-extension-shoot-networking-problemdetector-admission

WORKDIR /
COPY --from=builder /go/bin/gardener-extension-shoot-networking-problemdetector-admission /gardener-extension-shoot-networking-problemdetector-admission
ENTRYPOINT ["/gardener-extension-shoot-networking-problemdetector-admission"]
