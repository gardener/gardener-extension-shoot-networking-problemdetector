<p>Packages:</p>
<ul>
<li>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud%2fv1alpha1">shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1">shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the shoot networking filter extension configuration.</p>
</p>
Resource Types:
<ul><li>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration</a>
</li></ul>
<h3 id="shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration
</h3>
<p>
<p>Configuration contains information about the network problem detector configuration.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Configuration</code></td>
</tr>
<tr>
<td>
<code>networkProblemDetector</code></br>
<em>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.NetworkProblemDetector">
NetworkProblemDetector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NetworkProblemDetector contains the configuration for the network problem detector</p>
</td>
</tr>
<tr>
<td>
<code>healthCheckConfig</code></br>
<em>
<a href="https://github.com/gardener/gardener/extensions/pkg/apis/config">
github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1.HealthCheckConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HealthCheckConfig is the config for the health check controller.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.K8sExporter">K8sExporter
</h3>
<p>
(<em>Appears on:</em>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.NetworkProblemDetector">NetworkProblemDetector</a>)
</p>
<p>
<p>K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<p>Enabled if true, the K8s exporter is active</p>
</td>
</tr>
<tr>
<td>
<code>heartbeatPeriod</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#duration-v1-meta">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HeartbeatPeriod defines the update frequency of the node conditions.</p>
</td>
</tr>
<tr>
<td>
<code>minFailingPeerNodeShare</code></br>
<em>
float64
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinFailingPeerNodeShare if &gt; 0, reports node conditions <code>ClusterNetworkProblems</code> or <code>HostNetworkProblems</code> for node checks only if minimum share of destination peer nodes are failing. Valid range: [0.0,1.0]</p>
</td>
</tr>
</tbody>
</table>
<h3 id="shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.NetworkProblemDetector">NetworkProblemDetector
</h3>
<p>
(<em>Appears on:</em>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.Configuration">Configuration</a>)
</p>
<p>
<p>NetworkProblemDetector contains the configuration for the network problem detector.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>defaultPeriod</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#duration-v1-meta">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DefaultPeriod optionally overrides the default period for jobs running in the agent.</p>
</td>
</tr>
<tr>
<td>
<code>maxPeerNodes</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxPeerNodes optionally overrides the MaxPeerNodes in the agent config (maximum number of is the default period for jobs running in the agent.</p>
</td>
</tr>
<tr>
<td>
<code>pspDisabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>PSPDisabled is a flag to disable pod security policy.</p>
</td>
</tr>
<tr>
<td>
<code>pingEnabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>PingEnabled is a flag if ICMP ping checks should be performed.</p>
</td>
</tr>
<tr>
<td>
<code>k8sExporter</code></br>
<em>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1.K8sExporter">
K8sExporter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>K8sExporter configures the K8s exporter for updating node conditions and creating events.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
