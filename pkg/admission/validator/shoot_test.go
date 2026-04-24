// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator_test

import (
	"context"
	"encoding/json"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/admission/validator"
	configv1alpha1 "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

func marshalProviderConfig(cfg configv1alpha1.ShootProviderConfig) *runtime.RawExtension {
	raw, err := json.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())
	return &runtime.RawExtension{Raw: raw}
}

func newShoot(providerConfig *runtime.RawExtension) *gardencorev1beta1.Shoot {
	shoot := &gardencorev1beta1.Shoot{}
	shoot.Spec.Extensions = []gardencorev1beta1.Extension{
		{
			Type:           constants.ExtensionType,
			ProviderConfig: providerConfig,
		},
	}
	return shoot
}

var _ = Describe("ShootValidator", func() {
	var (
		ctx = context.Background()
		val = validator.NewShootValidator()
	)

	Describe("Validate", func() {
		It("returns nil when no extension of the correct type is present", func() {
			shoot := &gardencorev1beta1.Shoot{}
			shoot.Spec.Extensions = []gardencorev1beta1.Extension{{Type: "some-other-extension"}}
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil when the extension is present but ProviderConfig is nil", func() {
			shoot := newShoot(nil)
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil when ProviderConfig has no independentProbes", func() {
			cfg := configv1alpha1.ShootProviderConfig{}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil when pingEnabled is set but no independentProbes", func() {
			t := true
			cfg := configv1alpha1.ShootProviderConfig{PingEnabled: &t}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns error when ProviderConfig.Raw is invalid JSON", func() {
			shoot := newShoot(&runtime.RawExtension{Raw: []byte(`{invalid`)})
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when probe has empty jobID", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 80},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when probe has duplicate jobID", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 80},
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "other.com", Port: 443},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when TCP probe has neither host nor ipAddress", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Port: 80},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when TCP probe has port 0", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 0},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when TCP probe has port > 65535", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 65536},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when ipAddress is not a valid IP", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", IPAddress: "not-an-ip", Port: 80},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when HTTPS probe has no host", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolHTTPS, Port: 443},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when Ping probe has no ipAddress", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: configv1alpha1.ProbeProtocolPing, Host: "example.com"},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns error when probe has unknown protocol", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "probe1", Protocol: "UDP", Host: "example.com", Port: 53},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(HaveOccurred())
		})

		It("returns nil for a valid TCP probe", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "tcp-probe", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 443},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil for a valid TCP probe with ipAddress override", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "tcp-probe", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", IPAddress: "1.2.3.4", Port: 443},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil for a valid HTTPS probe", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "https-probe", Protocol: configv1alpha1.ProbeProtocolHTTPS, Host: "example.com", Port: 443},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil for a valid Ping probe", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "ping-probe", Protocol: configv1alpha1.ProbeProtocolPing, Host: "example.com", IPAddress: "1.2.3.4"},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns nil for multiple valid probes", func() {
			cfg := configv1alpha1.ShootProviderConfig{
				IndependentProbes: []configv1alpha1.IndependentProbe{
					{JobID: "tcp-probe", Protocol: configv1alpha1.ProbeProtocolTCP, Host: "example.com", Port: 80},
					{JobID: "https-probe", Protocol: configv1alpha1.ProbeProtocolHTTPS, Host: "api.example.com", Port: 443},
					{JobID: "ping-probe", Protocol: configv1alpha1.ProbeProtocolPing, IPAddress: "10.0.0.1"},
				},
			}
			shoot := newShoot(marshalProviderConfig(cfg))
			Expect(val.Validate(ctx, shoot, nil)).To(Succeed())
		})

		It("returns error for wrong object type", func() {
			Expect(val.Validate(ctx, &gardencorev1beta1.Seed{}, nil)).To(HaveOccurred())
		})
	})
})
