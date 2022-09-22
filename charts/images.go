// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"embed"
	_ "embed"
)

// Images YAML contains the contents of the images.yaml file.
//
//go:embed images.yaml
var ImagesYAML string

// Internal contains the internal charts
//
//go:embed internal
var Internal embed.FS

// ChartsPath is the path to the charts
const ChartsPath = "internal"
