// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"encoding/json"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	configv1alpha1 "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/validation"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

// shootProviderConfig mirrors the JSON shape of the per-shoot providerConfig.
// It is intentionally not a registered API type; plain JSON unmarshaling is used.
type shootProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	PingEnabled       *bool                             `json:"pingEnabled,omitempty"`
	IndependentProbes []configv1alpha1.IndependentProbe `json:"independentProbes,omitempty"`
}

type shoot struct{}

// NewShootValidator returns a new Validator for Shoot resources.
func NewShootValidator() extensionswebhook.Validator {
	return &shoot{}
}

// Validate validates the given Shoot object.
func (s *shoot) Validate(_ context.Context, newObj, _ client.Object) error {
	shootObj, ok := newObj.(*gardencorev1beta1.Shoot)
	if !ok {
		return fmt.Errorf("wrong object type %T", newObj)
	}

	for i, ext := range shootObj.Spec.Extensions {
		if ext.Type != constants.ExtensionType {
			continue
		}
		if ext.ProviderConfig == nil {
			return nil
		}
		fldPath := field.NewPath("spec", "extensions").Index(i).Child("providerConfig")
		return validateProviderConfig(ext.ProviderConfig, fldPath)
	}
	return nil
}

func validateProviderConfig(rawExt *runtime.RawExtension, fldPath *field.Path) error {
	if rawExt == nil || len(rawExt.Raw) == 0 {
		return nil
	}
	var cfg shootProviderConfig
	if err := json.Unmarshal(rawExt.Raw, &cfg); err != nil {
		return field.Invalid(fldPath, string(rawExt.Raw),
			fmt.Sprintf("failed to unmarshal providerConfig: %v", err))
	}
	if len(cfg.IndependentProbes) == 0 {
		return nil
	}
	probes := make([]config.IndependentProbe, len(cfg.IndependentProbes))
	for i, p := range cfg.IndependentProbes {
		probes[i] = config.IndependentProbe{
			JobID:     p.JobID,
			Protocol:  config.ProbeProtocol(p.Protocol),
			Host:      p.Host,
			IPAddress: p.IPAddress,
			Port:      p.Port,
			Period:    p.Period,
		}
	}
	if err := validation.ValidateIndependentProbes(probes); err != nil {
		return field.Invalid(fldPath.Child("independentProbes"), cfg.IndependentProbes, err.Error())
	}
	return nil
}
