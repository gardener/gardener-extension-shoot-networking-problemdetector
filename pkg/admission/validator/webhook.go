// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

const (
	// Name is the name of the validating webhook.
	Name = "shoot-networking-problemdetector-validator"

	// WebhookPath is the HTTP path at which the webhook is served.
	WebhookPath = "/webhooks/validate-shoot-networking-problemdetector"
)

// New creates a new validating webhook for Shoot resources carrying the
// shoot-networking-problemdetector extension.
func New(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	return extensionswebhook.New(mgr, extensionswebhook.Args{
		Name: Name,
		Path: WebhookPath,
		Validators: map[extensionswebhook.Validator][]extensionswebhook.Type{
			NewShootValidator(): {{Obj: &gardencorev1beta1.Shoot{}}},
		},
		Target: extensionswebhook.TargetSeed,
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"extensions.extensions.gardener.cloud/" + constants.ExtensionType: "true",
			},
		},
	})
}
