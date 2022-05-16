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

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/constants"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/imagevector"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"github.com/gardener/gardener/extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	"github.com/gardener/gardener/pkg/utils/chart"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/gardener/network-problem-detector/pkg/deploy"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// ActuatorName is the name of the Networking Policy Filter actuator.
	ActuatorName = constants.ServiceName + "-actuator"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(config config.Configuration) extension.Actuator {
	return &actuator{
		logger:        log.Log.WithName(ActuatorName),
		serviceConfig: config,
	}
}

type actuator struct {
	client        client.Client
	config        *rest.Config
	decoder       runtime.Decoder
	serviceConfig config.Configuration
	logger        logr.Logger
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	if !controller.IsHibernated(cluster) {
		if err := a.createShootResources(ctx, cluster, namespace); err != nil {
			return err
		}
	}

	if err := a.createSeedResources(ctx, cluster, namespace); err != nil {
		return err
	}

	return nil
}

func (a *actuator) createSeedResources(ctx context.Context, cluster *controller.Cluster, namespace string) error {
	values := map[string]interface{}{
		"genericTokenKubeconfigSecretName": extensions.GenericTokenKubeconfigSecretNameFromCluster(cluster),
		"shootClusterSecret":               gutil.SecretNamePrefixShootAccess + constants.ShootAccessSecretName,
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

	a.logger.Info("Component is being applied", "component", "network-problem-detector-controller", "namespace", namespace)

	return a.createManagedResource(ctx, namespace, constants.ManagedResourceNamesControllerSeed, "seed", renderer, constants.NetworkProblemDetectorControllerChartNameSeed, namespace, values, nil)
}

func (a *actuator) createShootResources(ctx context.Context, cluster *controller.Cluster, namespace string) error {
	defaultPeriod := 10 * time.Second
	pspEnabled := true
	pingEnabled := false
	if a.serviceConfig.NetworkProblemDetector != nil {
		if a.serviceConfig.NetworkProblemDetector.DefaultPeriod != nil {
			defaultPeriod = a.serviceConfig.NetworkProblemDetector.DefaultPeriod.Duration
		}
		if a.serviceConfig.NetworkProblemDetector.PSPDisabled != nil {
			pspEnabled = !*a.serviceConfig.NetworkProblemDetector.PSPDisabled
		}
	}

	shootResources, err := a.getShootAgentResources(defaultPeriod, pingEnabled, pspEnabled)
	if err != nil {
		return err
	}
	err = managedresources.CreateForShoot(ctx, a.client, namespace, constants.ManagedResourceNamesAgentShoot, false, shootResources)
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
	chartPath := filepath.Join(constants.ChartsPath, chartName)
	chart, err := renderer.Render(chartPath, chartName, chartNamespace, chartValues)
	if err != nil {
		return err
	}

	data := map[string][]byte{chartName: chart.Manifest()}
	keepObjects := false
	forceOverwriteAnnotations := false
	return managedresources.Create(ctx, a.client, namespace, name, false, class, data, &keepObjects, injectedLabels, &forceOverwriteAnnotations)
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	err := a.deleteShootResources(ctx, namespace)
	if err != nil {
		return err
	}

	return a.deleteSeedResources(ctx, namespace)
}

func (a *actuator) deleteShootResources(ctx context.Context, namespace string) error {
	a.logger.Info("Deleting managed resource for shoot", "namespace", namespace)
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

func (a *actuator) deleteSeedResources(ctx context.Context, namespace string) error {
	a.logger.Info("Deleting managed resource for seed", "namespace", namespace)

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
func (a *actuator) Restore(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, ex)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	// Keep objects for shoot managed resources so that they are not deleted from the shoot during the migration
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNamesAgentShoot, true); err != nil {
		return err
	}
	// Keep objects for shoot managed resources so that they are not deleted from the shoot during the migration
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNamesControllerShoot, true); err != nil {
		return err
	}

	return a.Delete(ctx, ex)
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

// InjectScheme injects the given scheme into the reconciler.
func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDecoder()
	return nil
}

func (a *actuator) getShootAgentResources(defaultPeriod time.Duration, pingEnabled, pspEnabled bool) (map[string][]byte, error) {
	shootRegistry := managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer)

	image, err := imagevector.ImageVector().FindImage(constants.AgentImageName)
	if err != nil {
		return nil, err
	}

	deployConfig := &deploy.AgentDeployConfig{
		Image:                    image.String(),
		DefaultPeriod:            defaultPeriod,
		PodSecurityPolicyEnabled: pspEnabled,
		PingEnabled:              pingEnabled,
	}
	objs, err := deploy.DeployNetworkProblemDetectorAgent(deployConfig)
	if err != nil {
		return nil, err
	}

	var objects []client.Object
	for _, obj := range objs {
		objects = append(objects, obj.(client.Object))
	}

	// clusterConfig is updated by nwpd controller later, but it is created here
	clusterConfig, err := deploy.BuildClusterConfig(nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	clusterCM, err := deploy.BuildClusterConfigMap(clusterConfig)
	addIgnoreAnnotation(clusterCM) // don't update
	objects = append(objects, clusterCM)

	agentConfig, err := deployConfig.BuildAgentConfig()
	if err != nil {
		return nil, err
	}
	agentCM, err := deploy.BuildAgentConfigMap(agentConfig)
	objects = append(objects, agentCM)

	shootResources, err := shootRegistry.AddAllAndSerialize(objects...)
	if err != nil {
		return nil, err
	}
	return shootResources, nil
}

func addIgnoreAnnotation(cm *corev1.ConfigMap) {
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[resourcesv1alpha1.Ignore] = "true"
}
