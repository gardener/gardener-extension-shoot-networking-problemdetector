// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "$TOOLS_BIN_DIR/extension-generator --name=extension-shoot-networking-problemdetector --provider-type=shoot-networking-problemdetector --component-category=extension --extension-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/shoot-networking-problemdetector:$(cat ../VERSION) --admission-runtime-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/admission-shoot-networking-problemdetector-runtime:$(cat ../VERSION) --admission-application-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/admission-shoot-networking-problemdetector-application:$(cat ../VERSION) --destination=./extension/base/extension.yaml"
//go:generate sh -c "$TOOLS_BIN_DIR/kustomize build ./extension -o ./extension.yaml"

package example
