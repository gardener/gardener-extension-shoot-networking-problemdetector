// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"os"

	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	apisconfig "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	controllerconfig "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/controller/config"
	healthcheckcontroller "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/controller/healthcheck"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/controller/lifecycle"
)

var (
	scheme  *runtime.Scheme
	decoder runtime.Decoder
)

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(apisconfig.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

// NetworkProblemDetectorOptions holds options related to the network problem detector controller.
type NetworkProblemDetectorOptions struct {
	ConfigLocation string
	config         *NetworkProblemDetectorConfig
}

// AddFlags implements Flagger.AddFlags.
func (o *NetworkProblemDetectorOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigLocation, "config", "", "Path to network problem detector configuration")
}

// Complete implements Completer.Complete.
func (o *NetworkProblemDetectorOptions) Complete() error {
	if o.ConfigLocation == "" {
		return errors.New("config location is not set")
	}
	data, err := os.ReadFile(o.ConfigLocation)
	if err != nil {
		return err
	}

	config := apisconfig.Configuration{}
	_, _, err = decoder.Decode(data, nil, &config)
	if err != nil {
		return err
	}

	o.config = &NetworkProblemDetectorConfig{
		config: config,
	}

	return nil
}

// Completed returns the decoded NetworkProblemDetectorConfig instance. Only call this if `Complete` was successful.
func (o *NetworkProblemDetectorOptions) Completed() *NetworkProblemDetectorConfig {
	return o.config
}

// NetworkProblemDetectorConfig contains configuration information about the network problem detector.
type NetworkProblemDetectorConfig struct {
	config apisconfig.Configuration
}

// Apply applies the NetworkProblemDetectorOptions to the passed ControllerOptions instance.
func (c *NetworkProblemDetectorConfig) Apply(config *controllerconfig.Config) {
	config.Configuration = c.config
}

// ApplyHealthCheckConfig applies the HealthCheckConfig to the config.
func (c *NetworkProblemDetectorConfig) ApplyHealthCheckConfig(config *healthcheckconfig.HealthCheckConfig) {
	if c.config.HealthCheckConfig != nil {
		*config = *c.config.HealthCheckConfig
	}
}

// ControllerSwitches are the cmd.ControllerSwitches for the extension controllers.
func ControllerSwitches() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(lifecycle.Name, lifecycle.AddToManager),
		cmd.Switch(extensionshealthcheckcontroller.ControllerName, healthcheckcontroller.AddToManager),
		cmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
