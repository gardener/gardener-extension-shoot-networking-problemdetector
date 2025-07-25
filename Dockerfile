# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.24.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-shoot-networking-problemdetector

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION
RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# gardener-extension-shoot-networking-problemdetector
FROM gcr.io/distroless/static-debian12:nonroot AS gardener-extension-shoot-networking-problemdetector

WORKDIR /
COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-shoot-networking-problemdetector /gardener-extension-shoot-networking-problemdetector
ENTRYPOINT ["/gardener-extension-shoot-networking-problemdetector"]
