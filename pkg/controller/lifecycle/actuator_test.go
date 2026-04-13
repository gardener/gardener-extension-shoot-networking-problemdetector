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

	Describe("independent probes", func() {
		Describe("#addIndependentProbeJobs", func() {
			It("adds TCP probe jobs to both HostNetwork and PodNetwork", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())
				hostJobsBefore := len(agentConfig.HostNetwork.Jobs)
				podJobsBefore := len(agentConfig.PodNetwork.Jobs)

				probes := []config.IndependentProbe{
					{JobID: "check-registry", Protocol: config.ProbeProtocolTCP, Host: "registry.example.com", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())
				Expect(agentConfig.HostNetwork.Jobs).To(HaveLen(hostJobsBefore + 1))
				Expect(agentConfig.PodNetwork.Jobs).To(HaveLen(podJobsBefore + 1))

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-h-check-registry"))
				Expect(hostJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "registry.example.com:registry.example.com:443"))

				podJob := agentConfig.PodNetwork.Jobs[len(agentConfig.PodNetwork.Jobs)-1]
				Expect(podJob.JobID).To(Equal("indep-p-check-registry"))
				Expect(podJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "registry.example.com:registry.example.com:443"))
			})

			It("uses IPAddress as the connection target for TCP probes when set", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "check-registry-ip", Protocol: config.ProbeProtocolTCP, Host: "registry.example.com", IPAddress: "192.0.2.1", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-h-check-registry-ip"))
				Expect(hostJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "registry.example.com:192.0.2.1:443"))
			})

			It("adds HTTPS probe jobs", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "check-api", Protocol: config.ProbeProtocolHTTPS, Host: "api.example.com", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-h-check-api"))
				Expect(hostJob.Args).To(ConsistOf("checkHTTPSGet", "--endpoints", "api.example.com:443"))
			})

			It("adds HTTPS probe jobs on non-standard port", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "check-http", Protocol: config.ProbeProtocolHTTPS, Host: "internal.example.com", Port: 8080},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-h-check-http"))
				Expect(hostJob.Args).To(ConsistOf("checkHTTPSGet", "--endpoints", "internal.example.com:8080"))
			})

			It("applies --period arg when Period is set", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				period := metav1.Duration{Duration: 30 * time.Second}
				probes := []config.IndependentProbe{
					{JobID: "check-slow", Protocol: config.ProbeProtocolTCP, Host: "slow.example.com", Port: 8080, Period: &period},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ContainElements("--period", "30s"))
			})

			It("returns error for duplicate jobID", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "dup", Protocol: config.ProbeProtocolTCP, Host: "a.example.com", Port: 80},
					{JobID: "dup", Protocol: config.ProbeProtocolTCP, Host: "b.example.com", Port: 80},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("duplicate independent probe jobID")))
			})

			It("uses IPAddress as endpoint label for TCP probes when host is omitted", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "check-ip-only", Protocol: config.ProbeProtocolTCP, IPAddress: "192.0.2.1", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ConsistOf("checkTCPPort", "--endpoints", "192.0.2.1:192.0.2.1:443"))
			})

			It("returns error for empty host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "bad", Protocol: config.ProbeProtocolTCP, Host: "", Port: 80},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("must have host or ipAddress")))
			})

			It("returns error for HTTPS probe without host", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "bad-https", Protocol: config.ProbeProtocolHTTPS, IPAddress: "192.0.2.1", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("requires host")))
			})

			It("returns error for out-of-range port", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "bad-port", Protocol: config.ProbeProtocolTCP, Host: "example.com", Port: 0},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("invalid port")))
			})

			It("returns error for invalid ipAddress", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "bad-ip", Protocol: config.ProbeProtocolTCP, Host: "example.com", IPAddress: "not-an-ip", Port: 443},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("invalid ipAddress")))
			})

			It("returns error for unsupported protocol", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "bad-proto", Protocol: config.ProbeProtocol("UDP"), Host: "example.com", Port: 53},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("adds Ping probe jobs using ipAddress", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "ping-gw", Protocol: config.ProbeProtocolPing, Host: "gateway.example.com", IPAddress: "192.0.2.1"},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.JobID).To(Equal("indep-h-ping-gw"))
				Expect(hostJob.Args).To(ConsistOf("pingHost", "--hosts", "gateway.example.com:192.0.2.1"))
			})

			It("adds Ping probe using ipAddress as label when host is omitted", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "ping-ip", Protocol: config.ProbeProtocolPing, IPAddress: "192.0.2.1"},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(Succeed())

				hostJob := agentConfig.HostNetwork.Jobs[len(agentConfig.HostNetwork.Jobs)-1]
				Expect(hostJob.Args).To(ConsistOf("pingHost", "--hosts", "192.0.2.1:192.0.2.1"))
			})

			It("returns error for Ping probe without ipAddress", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())

				probes := []config.IndependentProbe{
					{JobID: "ping-bad", Protocol: config.ProbeProtocolPing, Host: "gateway.example.com"},
				}
				Expect(addIndependentProbeJobs(agentConfig, probes)).To(MatchError(ContainSubstring("requires ipAddress")))
			})

			It("returns no error for empty probes list", func() {
				agentConfig, err := deployConfig.BuildAgentConfig()
				Expect(err).To(BeNil())
				Expect(addIndependentProbeJobs(agentConfig, nil)).To(Succeed())
			})
		})

		Describe("#decodeShootProviderConfig", func() {
			It("returns zero value for nil ProviderConfig", func() {
				cfg, err := decodeShootProviderConfig(nil)
				Expect(err).To(BeNil())
				Expect(cfg.IndependentProbes).To(BeNil())
				Expect(cfg.PingEnabled).To(BeNil())
			})

			It("decodes independent probes from raw JSON", func() {
				raw, err := json.Marshal(map[string]interface{}{
					"apiVersion": "shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1",
					"kind":       "NetworkProblemDetectorConfig",
					"independentProbes": []map[string]interface{}{
						{"jobID": "check-ext", "protocol": "HTTPS", "host": "ext.example.com", "port": 443},
					},
				})
				Expect(err).To(BeNil())

				cfg, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: raw})
				Expect(err).To(BeNil())
				Expect(cfg.IndependentProbes).To(HaveLen(1))
				Expect(cfg.IndependentProbes[0].JobID).To(Equal("check-ext"))
				Expect(cfg.IndependentProbes[0].Protocol).To(Equal(config.ProbeProtocolHTTPS))
				Expect(cfg.IndependentProbes[0].Host).To(Equal("ext.example.com"))
				Expect(cfg.IndependentProbes[0].Port).To(Equal(443))
			})

			It("decodes pingEnabled from raw JSON", func() {
				raw, err := json.Marshal(map[string]interface{}{
					"apiVersion":  "shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1",
					"kind":        "NetworkProblemDetectorConfig",
					"pingEnabled": true,
				})
				Expect(err).To(BeNil())

				cfg, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: raw})
				Expect(err).To(BeNil())
				Expect(cfg.PingEnabled).NotTo(BeNil())
				Expect(*cfg.PingEnabled).To(BeTrue())
			})

			It("returns error for invalid JSON", func() {
				_, err := decodeShootProviderConfig(&runtime.RawExtension{Raw: []byte("not-json")})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
