# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.18.2 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-shoot-networking-problemdetector
COPY . .
RUN make install

############# gardener-extension-shoot-networking-problemdetector
FROM alpine:3.15.4 AS gardener-extension-shoot-networking-problemdetector

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-shoot-networking-problemdetector /gardener-extension-shoot-networking-problemdetector
ENTRYPOINT ["/gardener-extension-shoot-networking-problemdetector"]
