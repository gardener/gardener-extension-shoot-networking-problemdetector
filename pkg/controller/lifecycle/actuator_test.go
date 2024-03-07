// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"time"

	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/network-problem-detector/pkg/common"
	"github.com/gardener/network-problem-detector/pkg/deploy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("activator methods", func() {
	var (
		deployConfig = &deploy.AgentDeployConfig{
			Image:         "image:tag",
			DefaultPeriod: 16 * time.Second,
			PingEnabled:   false,
		}
	)

	It("DeployNetworkProblemDetectorAgent", func() {
		objs, err := deploy.NetworkProblemDetectorAgent(deployConfig)
		Expect(err).To(BeNil())
		Expect(len(objs)).NotTo(BeZero())
		var ds *appsv1.DaemonSet
	loop:
		for _, obj := range objs {
			switch v := obj.(type) {
			case *appsv1.DaemonSet:
				ds = v
				break loop
			}
		}
		Expect(ds).NotTo(BeNil())
		Expect(len(ds.Spec.Template.Spec.Containers)).To(Equal(1))
		Expect(ds.Spec.Template.Spec.Containers[0].Image).To(Equal("image:tag"))
	})

	It("#buildDefaultClusterConfigMap works", func() {
		cm, err := buildDefaultClusterConfigMap()
		Expect(err).To(BeNil())
		Expect(cm).NotTo(BeNil())
		Expect(cm.Annotations[resourcesv1alpha1.Ignore]).To(Equal("true"))
		data, ok := cm.Data[common.ClusterConfigFilename]
		Expect(ok).To(BeTrue())
		Expect(len(data)).NotTo(BeZero())
	})

	It("#buildDefaultAgentConfigMap works", func() {
		cm, err := buildDefaultAgentConfigMap(deployConfig)
		Expect(err).To(BeNil())
		Expect(cm).NotTo(BeNil())
		data, ok := cm.Data[common.AgentConfigFilename]
		Expect(ok).To(BeTrue())
		Expect(len(data)).NotTo(BeZero())
	})
})
