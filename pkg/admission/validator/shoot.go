// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"errors"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigsjson "sigs.k8s.io/json"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	configv1alpha1 "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/validation"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

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
	var cfg configv1alpha1.ShootProviderConfig
	strictErrs, err := sigsjson.UnmarshalStrict(rawExt.Raw, &cfg)
	if err != nil || len(strictErrs) > 0 {
		return field.Invalid(fldPath, string(rawExt.Raw),
			fmt.Sprintf("failed to decode providerConfig: %v", errors.Join(append(strictErrs, err)...)))
	}
	if len(cfg.AdditionalProbes) == 0 {
		return nil
	}
	probes := make([]config.AdditionalProbe, len(cfg.AdditionalProbes))
	for i, p := range cfg.AdditionalProbes {
		probes[i] = config.AdditionalProbe{
			JobID:    p.JobID,
			Protocol: config.ProbeProtocol(p.Protocol),
			Host:     p.Host,
			Port:     p.Port,
			Period:   p.Period,
		}
	}
	if err := validation.ValidateAdditionalProbes(probes); err != nil {
		return field.Invalid(fldPath.Child("additionalProbes"), cfg.AdditionalProbes, err.Error())
	}
	return nil
}
