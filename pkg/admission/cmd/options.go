// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	extensionscmdwebhook "github.com/gardener/gardener/extensions/pkg/webhook/cmd"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/admission/validator"
)

// GardenWebhookSwitchOptions are the extensionscmdwebhook.SwitchOptions for the admission webhooks.
func GardenWebhookSwitchOptions() *extensionscmdwebhook.SwitchOptions {
	return extensionscmdwebhook.NewSwitchOptions(
		extensionscmdwebhook.Switch(validator.Name, validator.New),
	)
}
