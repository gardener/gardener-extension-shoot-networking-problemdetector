// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"encoding/json"
	"time"

	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/network-problem-detector/pkg/common"
	"github.com/gardener/network-problem-detector/pkg/deploy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
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

	It("#buildDefaultAgentConfigMap works without probes", func() {
		cm, err := buildDefaultAgentConfigMap(deployConfig, nil)
		Expect(err).To(BeNil())
		Expect(cm).NotTo(BeNil())
		data, ok := cm.Data[common.AgentConfigFilename]
		Expect(ok).To(BeTrue())
		Expect(len(data)).NotTo(BeZero())
	})

	Describe("additional probes", func() {
		Describe("#addAdditionalProbeJobs", func() {
			It("adds TCP probe jobs to both HostNetwork and PodNetwork", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())
				hostJobsBefore := len(agentConfig.HostNetwork.Jobs)
				podJobsBefore := len(agentConfig.PodNetwork.Jobs)

				probes := []config.AdditionalProbe{
					{JobID: "check-registry", Protocol: config.ProbeProtocolTCP, Host: "registry.example.com", Port: 443},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())
				Expect(agentConfig.HostNetwork.Jobs).To(HaveLen(hostJobsBefore + 1))
				Expect(agentConfig.PodNetwork.Jobs).To(HaveLen(podJobsBefore + 1))

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-tcp-n-check-registry"))
				Expect(hostJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "registry.example.com:registry.example.com:443", "--period", "60s"))

				podJob := agentConfig.PodNetwork.Jobs[len(agentConfig.PodNetwork.Jobs)-1]
				Expect(podJob.JobID).To(Equal("indep-tcp-p-check-registry"))
				Expect(podJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "registry.example.com:registry.example.com:443", "--period", "60s"))
			})

			It("adds TCP probe jobs using an IP address as host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "check-ip", Protocol: config.ProbeProtocolTCP, Host: "192.0.2.1", Port: 443},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "192.0.2.1:192.0.2.1:443", "--period", "60s"))
			})

			It("adds HTTPS probe jobs", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "check-api", Protocol: config.ProbeProtocolHTTPS, Host: "api.example.com", Port: 443},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-https-n-check-api"))
				Expect(hostJob.Args).To(ConsistOf("checkHTTPSGet", "--endpoints", "api.example.com:443", "--period", "60s"))
			})

			It("adds HTTPS probe jobs on non-standard port", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "check-http", Protocol: config.ProbeProtocolHTTPS, Host: "internal.example.com", Port: 8080},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-https-n-check-http"))
				Expect(hostJob.Args).To(ConsistOf("checkHTTPSGet", "--endpoints", "internal.example.com:8080", "--period", "60s"))
			})

			It("applies --period arg when Period is set", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				period := metav1.Duration{Duration: 120 * time.Second}
				probes := []config.AdditionalProbe{
					{JobID: "check-slow", Protocol: config.ProbeProtocolTCP, Host: "slow.example.com", Port: 8080, Period: &period},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ContainElements("--period", "2m0s"))
			})

			It("returns error for duplicate jobID", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "dup", Protocol: config.ProbeProtocolTCP, Host: "a.example.com", Port: 80},
					{JobID: "dup", Protocol: config.ProbeProtocolTCP, Host: "b.example.com", Port: 80},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("duplicate additional probe jobID")))
			})

			It("returns error for empty host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "bad", Protocol: config.ProbeProtocolTCP, Host: "", Port: 80},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("must have host")))
			})

			It("returns error for HTTPS probe without host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "bad-https", Protocol: config.ProbeProtocolHTTPS, Port: 443},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("must have host")))
			})

			It("returns error for out-of-range port", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "bad-port", Protocol: config.ProbeProtocolTCP, Host: "example.com", Port: 0},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("invalid port")))
			})

			It("returns error for unsupported protocol", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "bad-proto", Protocol: config.ProbeProtocol("UDP"), Host: "example.com", Port: 53},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("adds ICMP probe jobs using hostname", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "ping-gw", Protocol: config.ProbeProtocolICMP, Host: "gateway.example.com"},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-icmp-n-ping-gw"))
				Expect(hostJob.Args).To(ConsistOf("pingHost", "--hosts", "gateway.example.com:gateway.example.com", "--period", "60s"))
			})

			It("adds ICMP probe jobs using IP address", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "ping-ip", Protocol: config.ProbeProtocolICMP, Host: "192.0.2.1"},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ConsistOf("pingHost", "--hosts", "192.0.2.1:192.0.2.1", "--period", "60s"))
			})

			It("returns error for ICMP probe without host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.AdditionalProbe{
					{JobID: "icmp-bad", Protocol: config.ProbeProtocolICMP},
				}
				Expect(addAdditionalProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("must have host")))
			})

			It("returns no error for empty probes list", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())
				Expect(addAdditionalProbeJobs(agentConfig, nil)).To(Succeed())
			})
		})

		Describe("#decodeShootProviderConfig", func() {
			It("returns zero value for nil ProviderConfig", func() {
				cfg, err := decodeShootProviderConfig(nil)
				Expect(err).To(BeNil())
				Expect(cfg.AdditionalProbes).To(BeNil())
				Expect(cfg.IcmpEnabled).To(BeNil())
			})

			It("decodes additional probes from raw JSON", func() {
				raw, err := json.Marshal(map[string]interface{}{
					"apiVersion": "shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1",
					"kind":       "NetworkProblemDetectorConfig",
					"additionalProbes": []map[string]interface{}{
						{"jobID": "check-ext", "protocol": "HTTPS", "host": "ext.example.com", "port": 443},
					},
				})
				Expect(err).To(BeNil())

				cfg, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: raw})
				Expect(err).To(BeNil())
				Expect(cfg.AdditionalProbes).To(HaveLen(1))
				Expect(cfg.AdditionalProbes[0].JobID).To(Equal("check-ext"))
				Expect(cfg.AdditionalProbes[0].Protocol).To(Equal(config.ProbeProtocolHTTPS))
				Expect(cfg.AdditionalProbes[0].Host).To(Equal("ext.example.com"))
				Expect(cfg.AdditionalProbes[0].Port).To(Equal(443))
			})

			It("decodes icmpEnabled from raw JSON", func() {
				raw, err := json.Marshal(map[string]interface{}{
					"apiVersion":  "shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1",
					"kind":        "NetworkProblemDetectorConfig",
					"icmpEnabled": true,
				})
				Expect(err).To(BeNil())

				cfg, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: raw})
				Expect(err).To(BeNil())
				Expect(cfg.IcmpEnabled).NotTo(BeNil())
				Expect(*cfg.IcmpEnabled).To(BeTrue())
			})

			It("returns error for invalid JSON", func() {
				_, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: []byte("not-json")})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
