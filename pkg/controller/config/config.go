// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
)

// Config contains configuration for the network problem detector.
type Config struct {
	config.Configuration
}
