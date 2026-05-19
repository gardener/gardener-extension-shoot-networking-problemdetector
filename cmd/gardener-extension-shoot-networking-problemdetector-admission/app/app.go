// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"

	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionscmdwebhook "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	gardencoreinstall "github.com/gardener/gardener/pkg/apis/core/install"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	admissioncmd "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/admission/cmd"
)

const admissionName = "networking-problemdetector-admission"

var log = logf.Log.WithName("gardener-extension-" + admissionName)

// NewAdmissionCommand creates a new command for running the shoot-networking-problemdetector admission webhook.
func NewAdmissionCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &extensionscmdcontroller.RESTOptions{}
		mgrOpts  = &extensionscmdcontroller.ManagerOptions{
			LeaderElection:     true,
			LeaderElectionID:   extensionscmdcontroller.LeaderElectionNameID(admissionName),
			WebhookServerPort:  443,
			MetricsBindAddress: ":8080",
			HealthBindAddress:  ":8081",
			WebhookCertDir:     "/tmp/admission-shoot-networking-problemdetector-cert",
		}

		webhookServerOptions = &extensionscmdwebhook.ServerOptions{
			Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
		}
		webhookSwitches = admissioncmd.GardenWebhookSwitchOptions()
		webhookOptions  = extensionscmdwebhook.NewAddToManagerOptions(
			admissionName,
			"",
			nil,
			nil,
			webhookServerOptions,
			webhookSwitches,
		)

		aggOption = extensionscmdcontroller.NewOptionAggregator(
			restOpts,
			mgrOpts,
			webhookOptions,
		)
	)

	cmd := &cobra.Command{
		Use:           "gardener-extension-shoot-networking-problemdetector-admission",
		Short:         "Admission webhook for the shoot-networking-problemdetector extension.",
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			verflag.PrintAndExitIfRequested()

			log.Info("Starting "+admissionName, "version", version.Get())

			if gardenKubeconfig := os.Getenv("GARDEN_KUBECONFIG"); gardenKubeconfig != "" {
				log.Info("Getting rest config for garden from GARDEN_KUBECONFIG", "path", gardenKubeconfig)
				restOpts.Kubeconfig = gardenKubeconfig
			}

			if err := aggOption.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			cmd.SilenceUsage = true

			managerOptions := mgrOpts.Completed().Options()

			// When SOURCE_CLUSTER is set, the manager connects to the virtual garden cluster
			// (where Shoot is defined) while the source cluster handles cert management in the
			// runtime cluster.
			var sourceClusterConfig *rest.Config
			if sourceClusterEnabled := os.Getenv("SOURCE_CLUSTER"); sourceClusterEnabled != "" {
				log.Info("Configuring source cluster option")
				var err error
				sourceClusterConfig, err = clientcmd.BuildConfigFromFlags("", os.Getenv("SOURCE_KUBECONFIG"))
				if err != nil {
					return err
				}
				managerOptions.LeaderElectionConfig = sourceClusterConfig
			} else {
				// Restrict secret cache to the webhook namespace to avoid cluster-wide permissions.
				managerOptions.Cache = cache.Options{
					ByObject: map[client.Object]cache.ByObject{
						&corev1.Secret{}: {Namespaces: map[string]cache.Config{webhookOptions.Server.Completed().Namespace: {}}},
					},
				}
			}

			mgr, err := manager.New(restOpts.Completed().Config, managerOptions)
			if err != nil {
				return fmt.Errorf("could not instantiate manager: %w", err)
			}

			gardencoreinstall.Install(mgr.GetScheme())

			var sourceCluster cluster.Cluster
			if sourceClusterConfig != nil {
				sourceCluster, err = cluster.New(sourceClusterConfig, func(opts *cluster.Options) {
					opts.Logger = log
					opts.Cache.DefaultNamespaces = map[string]cache.Config{v1beta1constants.GardenNamespace: {}}
				})
				if err != nil {
					return err
				}

				if err := mgr.AddReadyzCheck("source-informer-sync", gardenerhealthz.NewCacheSyncHealthz(sourceCluster.GetCache())); err != nil {
					return err
				}

				if err = mgr.Add(sourceCluster); err != nil {
					return err
				}
			}

			log.Info("Setting up webhook server")
			if _, err := webhookOptions.Completed().AddToManager(ctx, mgr, sourceCluster); err != nil {
				return fmt.Errorf("could not add webhook to manager: %w", err)
			}

			if err := mgr.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
				return fmt.Errorf("could not add ready check for informers: %w", err)
			}

			if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
				return fmt.Errorf("could not add health check to manager: %w", err)
			}

			if err := mgr.AddReadyzCheck("webhook-server", mgr.GetWebhookServer().StartedChecker()); err != nil {
				return fmt.Errorf("could not add ready check for webhook server to manager: %w", err)
			}

			return mgr.Start(ctx)
		},
	}

	verflag.AddFlags(cmd.Flags())
	aggOption.AddFlags(cmd.Flags())

	return cmd
}
