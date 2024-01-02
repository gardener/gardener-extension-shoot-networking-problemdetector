#!/bin/bash

# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

GARDENER_HACK_DIR=$(go list -m -f '{{.Dir}}' github.com/gardener/gardener)/hack

bash $(GARDENER_HACK_DIR)/hook-me.sh gardener-extension-shoot-networking-problemdetector $@
