// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

/**
	Overview
		- Tests the health checks for the shoot-networking-problemdetector extension.
	Prerequisites
		- A Shoot exists.
	Test-case:
		1) Extension CRD
			1.1) HealthCondition Type: ShootSystemComponentsHealthy
				-  update the ManagedResource 'extension-shoot-networking-problemdetector-shoot' and verify the health check conditions in the Extension CRD status.
 **/

package healthcheck

import (
	"context"
	"fmt"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/test/framework"
	healthcheckoperation "github.com/gardener/gardener/test/testmachinery/extensions/healthcheck"
	"github.com/onsi/ginkgo/v2"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

const (
	timeout = 5 * time.Minute
)

var _ = ginkgo.Describe("Extension-shoot-networking-problemdetector integration test: health checks", func() {
	f := framework.NewShootFramework(nil)

	ginkgo.Context("Extension", func() {
		ginkgo.Context("Condition type: ShootSystemComponentsHealthy", func() {
			f.Serial().Release().CIt(fmt.Sprintf("Extension CRD should contain unhealthy condition due to ManagedResource '%s' is unhealthy", constants.ManagedResourceNamesAgentShoot), func(ctx context.Context) {
				err := healthcheckoperation.ExtensionHealthCheckWithManagedResource(ctx, timeout, f, "shoot-networking-problemdetector", constants.ManagedResourceNamesControllerSeed, gardencorev1beta1.ShootObservabilityComponentsHealthy)
				framework.ExpectNoError(err)
			}, timeout)
		})
	})
})
