// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"github.com/gardener/gardener/extensions/pkg/util"
	corev1betaconstants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	"github.com/gardener/gardener/pkg/utils/chart"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/gardener/network-problem-detector/pkg/common"
	"github.com/gardener/network-problem-detector/pkg/deploy"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/charts"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/imagevector"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(mgr manager.Manager, config config.Configuration) extension.Actuator {
	return &actuator{
		client:        mgr.GetClient(),
		config:        mgr.GetConfig(),
		decoder:       serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
		serviceConfig: config,
	}
}

type actuator struct {
	client        client.Client
	config        *rest.Config
	decoder       runtime.Decoder
	serviceConfig config.Configuration
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	if !controller.IsHibernated(cluster) {
		if err := a.createShootResources(ctx, cluster, namespace, gardencorev1beta1helper.IsPSPDisabled(cluster.Shoot)); err != nil {
			return err
		}
	}

	if err := a.createSeedResources(ctx, log, cluster, namespace); err != nil {
		return err
	}

	return nil
}

func (a *actuator) createSeedResources(ctx context.Context, log logr.Logger, cluster *controller.Cluster, namespace string) error {
	values := map[string]interface{}{
		"replicaCount":                     controller.GetReplicas(cluster, 1),
		"genericTokenKubeconfigSecretName": extensions.GenericTokenKubeconfigSecretNameFromCluster(cluster),
		"shootClusterSecret":               gutil.SecretNamePrefixShootAccess + constants.ShootAccessSecretName,
		"priorityClassName":                corev1betaconstants.PriorityClassNameShootControlPlane200,
	}

	if err := gutil.NewShootAccessSecret(constants.ShootAccessSecretName, namespace).Reconcile(ctx, a.client); err != nil {
		return err
	}

	values, err := chart.InjectImages(values, imagevector.ImageVector(), []string{constants.ControllerImageName})
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", constants.ControllerImageName, err)
	}

	renderer, err := chartrenderer.NewForConfig(a.config)
	if err != nil {
		return fmt.Errorf("could not create chart renderer: %w", err)
	}

	log.Info("Component is being applied", "component", "network-problem-detector-controller", "namespace", namespace)

	return a.createManagedResource(ctx, namespace, constants.ManagedResourceNamesControllerSeed, "seed", renderer, constants.NetworkProblemDetectorControllerChartNameSeed, namespace, values, nil)
}

func (a *actuator) createShootResources(ctx context.Context, cluster *controller.Cluster, namespace string, pspDisabled bool) error {
	defaultPeriod := 5 * time.Second
	maxPeerNodes := 25
	pspDisabledByConfig := false
	pingEnabled := false
	var k8sExporter *config.K8sExporter
	if a.serviceConfig.NetworkProblemDetector != nil {
		if a.serviceConfig.NetworkProblemDetector.DefaultPeriod != nil {
			defaultPeriod = a.serviceConfig.NetworkProblemDetector.DefaultPeriod.Duration
		}
		if a.serviceConfig.NetworkProblemDetector.MaxPeerNodes != nil {
			maxPeerNodes = *a.serviceConfig.NetworkProblemDetector.MaxPeerNodes
		}
		if a.serviceConfig.NetworkProblemDetector.PSPDisabled != nil {
			pspDisabledByConfig = *a.serviceConfig.NetworkProblemDetector.PSPDisabled
		}
		if a.serviceConfig.NetworkProblemDetector.PingEnabled != nil {
			pingEnabled = !*a.serviceConfig.NetworkProblemDetector.PingEnabled
		}
		if a.serviceConfig.NetworkProblemDetector.K8sExporter != nil && a.serviceConfig.NetworkProblemDetector.K8sExporter.Enabled {
			k8sExporter = a.serviceConfig.NetworkProblemDetector.K8sExporter
		}
	}

	shootResources, err := a.getShootAgentResources(defaultPeriod, pingEnabled, !pspDisabled && !pspDisabledByConfig, k8sExporter, maxPeerNodes)
	if err != nil {
		return err
	}
	err = managedresources.CreateForShoot(ctx, a.client, namespace, constants.ManagedResourceNamesAgentShoot, "gardener-extension-shoot-networking-problemdetector", false, shootResources)
	if err != nil {
		return err
	}

	values := map[string]interface{}{
		"kubernetesVersion":             cluster.Shoot.Spec.Kubernetes.Version,
		"shootAccessServiceAccountName": constants.ShootAccessServiceAccountName,
	}

	renderer, err := util.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("could not create chart renderer: %w", err)
	}

	return a.createManagedResource(ctx, namespace, constants.ManagedResourceNamesControllerShoot, "", renderer, constants.NetworkProblemDetectorControllerChartNameShoot, metav1.NamespaceSystem, values, nil)
}

func (a *actuator) createManagedResource(ctx context.Context, namespace, name, class string, renderer chartrenderer.Interface, chartName, chartNamespace string, chartValues map[string]interface{}, injectedLabels map[string]string) error {
	chartPath := filepath.Join(charts.ChartsPath, chartName)
	chart, err := renderer.RenderEmbeddedFS(charts.Internal, chartPath, chartName, chartNamespace, chartValues)
	if err != nil {
		return err
	}

	data := map[string][]byte{chartName: chart.Manifest()}
	keepObjects := false
	forceOverwriteAnnotations := false
	return managedresources.Create(ctx, a.client, namespace, name, nil, false, class, data, &keepObjects, injectedLabels, &forceOverwriteAnnotations)
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	err := a.deleteShootResources(ctx, log, namespace)
	if err != nil {
		return err
	}

	return a.deleteSeedResources(ctx, log, namespace)
}

// ForceDelete the Extension resource.
func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, log, ex)
}

func (a *actuator) deleteShootResources(ctx context.Context, log logr.Logger, namespace string) error {
	log.Info("Deleting managed resource for shoot", "namespace", namespace)
	if err := managedresources.DeleteForShoot(ctx, a.client, namespace, constants.ManagedResourceNamesAgentShoot); err != nil {
		return err
	}
	if err := managedresources.DeleteForShoot(ctx, a.client, namespace, constants.ManagedResourceNamesControllerShoot); err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, constants.ManagedResourceNamesAgentShoot); err != nil {
		return err
	}
	if err := managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, constants.ManagedResourceNamesControllerShoot); err != nil {
		return err
	}
	return nil
}

func (a *actuator) deleteSeedResources(ctx context.Context, log logr.Logger, namespace string) error {
	log.Info("Deleting managed resource for seed", "namespace", namespace)

	if err := kutil.DeleteObject(ctx, a.client, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: gutil.SecretNamePrefixShootAccess + constants.ShootAccessSecretName, Namespace: namespace}}); err != nil {
		return err
	}

	if err := managedresources.Delete(ctx, a.client, namespace, constants.ManagedResourceNamesControllerSeed, false); err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return managedresources.WaitUntilDeleted(timeoutCtx, a.client, namespace, constants.ManagedResourceNamesControllerSeed)
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, log, ex)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Keep objects for shoot managed resources so that they are not deleted from the shoot during the migration
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNamesAgentShoot, true); err != nil {
		return err
	}
	// Keep objects for shoot managed resources so that they are not deleted from the shoot during the migration
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNamesControllerShoot, true); err != nil {
		return err
	}

	return a.Delete(ctx, log, ex)
}

func (a *actuator) getShootAgentResources(defaultPeriod time.Duration, pingEnabled, pspEnabled bool, k8sExporter *config.K8sExporter, maxPeerNodes int) (map[string][]byte, error) {
	shootRegistry := managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer)

	image, err := imagevector.ImageVector().FindImage(constants.AgentImageName)
	if err != nil {
		return nil, err
	}

	deployConfig := &deploy.AgentDeployConfig{
		Image:                        image.String(),
		DefaultPeriod:                defaultPeriod,
		MaxPeerNodes:                 maxPeerNodes,
		DefaultSeccompProfileEnabled: true,
		PodSecurityPolicyEnabled:     pspEnabled,
		PingEnabled:                  pingEnabled,
		PriorityClassName:            corev1betaconstants.PriorityClassNameShootSystem900,
		AdditionalLabels: map[string]string{
			"networking.gardener.cloud/to-apiserver":        "allowed",
			"networking.gardener.cloud/to-dns":              "allowed",
			"networking.gardener.cloud/to-from-nwpd-agents": "allowed",
		},
		// projected service account token is provided by the resource manager
		DisableAutomountServiceAccountTokenForAgents: true,
	}
	if k8sExporter != nil && k8sExporter.Enabled {
		deployConfig.K8sExporterEnabled = true
		deployConfig.K8sExporterHeartbeat = 3 * time.Minute
		if k8sExporter.HeartbeatPeriod != nil {
			if k8sExporter.HeartbeatPeriod.Duration < 1*time.Minute {
				return nil, fmt.Errorf("Invalid k8sExporter.heartbeatPeriod. Must be >= 1m")
			}
			deployConfig.K8sExporterHeartbeat = k8sExporter.HeartbeatPeriod.Duration
		}
		if k8sExporter.MinFailingPeerNodeShare != nil {
			deployConfig.K8sExporterMinFailingPeerNodeShare = *k8sExporter.MinFailingPeerNodeShare
		}
	}
	objs, err := deploy.DeployNetworkProblemDetectorAgent(deployConfig)
	if err != nil {
		return nil, err
	}

	var objects []client.Object
	for _, obj := range objs {
		objects = append(objects, obj.(client.Object))
	}

	networkPolicy := buildAgentNetworkPolicy()
	objects = append(objects, networkPolicy)

	clusterCM, err := buildDefaultClusterConfigMap()
	if err != nil {
		return nil, err
	}
	objects = append(objects, clusterCM)

	agentCM, err := buildDefaultAgentConfigMap(deployConfig)
	if err != nil {
		return nil, err
	}
	objects = append(objects, agentCM)

	shootResources, err := shootRegistry.AddAllAndSerialize(objects...)
	if err != nil {
		return nil, err
	}
	return shootResources, nil
}

func buildAgentNetworkPolicy() client.Object {
	tcp := corev1.ProtocolTCP
	podHttpPort := intstr.FromInt(common.PodNetPodHttpPort)
	hostHttpPort := intstr.FromInt(common.HostNetPodHttpPort)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gardener.cloud--allow-to-from-nwpd-agents",
			Namespace: constants.NamespaceKubeSystem,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"networking.gardener.cloud/to-from-nwpd-agents": "allowed",
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &tcp,
							Port:     &podHttpPort,
						},
					},
					To: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"gardener.cloud/role": "network-problem-detector"},
							},
						},
					},
				},
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &tcp,
							Port:     &hostHttpPort,
						},
					},
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &tcp,
							Port:     &podHttpPort,
						},
						{
							Protocol: &tcp,
							Port:     &hostHttpPort,
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
		},
	}
}

func buildDefaultAgentConfigMap(deployConfig *deploy.AgentDeployConfig) (*corev1.ConfigMap, error) {
	agentConfig, err := deployConfig.BuildAgentConfig()
	if err != nil {
		return nil, err
	}
	agentCM, err := deploy.BuildAgentConfigMap(agentConfig)
	if err != nil {
		return nil, err
	}
	return agentCM, nil
}

func buildDefaultClusterConfigMap() (*corev1.ConfigMap, error) {
	// clusterConfig is updated by nwpd controller later, but it is created here
	clusterConfig, err := deploy.BuildClusterConfig(nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	clusterCM, err := deploy.BuildClusterConfigMap(clusterConfig)
	if err != nil {
		return nil, err
	}
	addIgnoreAnnotation(clusterCM) // don't update
	return clusterCM, nil
}

func addIgnoreAnnotation(cm *corev1.ConfigMap) {
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[resourcesv1alpha1.Ignore] = "true"
}
