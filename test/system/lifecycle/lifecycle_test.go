// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

/**
	Overview
		- Tests the lifecycle controller for the shoot-networking-problemdetector extension.
	Prerequisites
		- A Shoot exists and the shoot-networking-problemdetector extension is available for the seed cluster.
**/

package lifecycle_test

import (
	"context"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/utils/test/matchers"
	"github.com/gardener/gardener/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

func init() {
	framework.RegisterShootFrameworkFlags()
}

const (
	timeout = 30 * time.Minute
)

var _ = Describe("Shoot networking problem detector testing", func() {
	f := framework.NewShootFramework(nil)

	var initialExtensionConfig []gardencorev1beta1.Extension

	BeforeEach(func() {
		initialExtensionConfig = f.Shoot.Spec.Extensions
	})

	AfterEach(func() {
		// Revert to initial extension configuration
		_ = f.UpdateShoot(context.Background(), func(shoot *gardencorev1beta1.Shoot) error {
			shoot.Spec.Extensions = initialExtensionConfig
			return nil
		})
	})

	f.Serial().Beta().CIt("Should perform the common case scenario without any errors", func(ctx context.Context) {
		err := f.UpdateShoot(ctx, ensureShootNetworkingFilterIsEnabled)
		Expect(err).ToNot(HaveOccurred())

		ds := &appsv1.DaemonSet{}
		// Verify that the egress filter applier daemonset exists
		err = f.ShootClient.Client().Get(ctx, client.ObjectKey{Namespace: constants.NamespaceKubeSystem, Name: constants.ApplicationName}, ds)
		Expect(err).ToNot(HaveOccurred())

		// Ensure that the networking filter is disabled in order to verify the deletion process
		err = f.UpdateShoot(ctx, ensureShootNetworkingFilterIsDisabled)
		Expect(err).NotTo(HaveOccurred())

		// Verify that the egress filter applier daemonset does not exist
		err = f.ShootClient.Client().Get(ctx, client.ObjectKey{Namespace: constants.NamespaceKubeSystem, Name: constants.ApplicationName}, ds)
		Expect(err).To(HaveOccurred())
		Expect(err).To(matchers.BeNotFoundError())
	}, timeout)
})

func ensureShootNetworkingFilterIsEnabled(shoot *gardencorev1beta1.Shoot) error {
	for i, e := range shoot.Spec.Extensions {
		if e.Type == constants.ExtensionType {
			if e.Disabled != nil && *e.Disabled == true {
				shoot.Spec.Extensions[i].Disabled = pointer.Bool(false)
			}
			return nil
		}
	}

	shoot.Spec.Extensions = append(shoot.Spec.Extensions, gardencorev1beta1.Extension{
		Type:     constants.ExtensionType,
		Disabled: pointer.Bool(false),
	})
	return nil
}

func ensureShootNetworkingFilterIsDisabled(shoot *gardencorev1beta1.Shoot) error {
	for i, e := range shoot.Spec.Extensions {
		if e.Type == constants.ExtensionType {
			shoot.Spec.Extensions[i].Disabled = pointer.Bool(true)
			return nil
		}
	}
	shoot.Spec.Extensions = append(shoot.Spec.Extensions, gardencorev1beta1.Extension{
		Type:     constants.ExtensionType,
		Disabled: pointer.Bool(true),
	})
	return nil
}
