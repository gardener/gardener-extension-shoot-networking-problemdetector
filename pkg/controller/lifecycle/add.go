// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
	controllerconfig "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/controller/config"
)

const (
	// Type is the type of Extension resource.
	Type = constants.ExtensionType
	// Name is the name of the lifecycle controller.
	Name = "shoot_networking_problemdetector_lifecycle_controller"
	// FinalizerSuffix is the finalizer suffix for the network problem detector controller.
	FinalizerSuffix = constants.ExtensionType
)

// DefaultAddOptions contains configuration for the network problem detector.
var DefaultAddOptions = AddOptions{}

// AddOptions are options to apply when adding the policy filter controller to the manager.
type AddOptions struct {
	// ControllerOptions contains options for the controller.
	ControllerOptions controller.Options
	// ServiceConfig contains configuration for the network problem detector.
	ServiceConfig controllerconfig.Config
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManager adds a Networking Policy Filter Lifecycle controller to the given Controller Manager.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return extension.Add(ctx, mgr, extension.AddArgs{
		Actuator:          NewActuator(mgr, DefaultAddOptions.ServiceConfig.Configuration),
		ControllerOptions: DefaultAddOptions.ControllerOptions,
		Name:              Name,
		FinalizerSuffix:   FinalizerSuffix,
		Resync:            60 * time.Minute,
		Predicates:        extension.DefaultPredicates(ctx, mgr, DefaultAddOptions.IgnoreOperationAnnotation),
		Type:              constants.ExtensionType,
	})
}
