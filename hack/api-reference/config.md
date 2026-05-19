<p>Packages:</p>
<ul>
<li>
<a href="#shoot-networking-problemdetector.extensions.config.gardener.cloud%2fv1alpha1">shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1">shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="additionalprobe">AdditionalProbe
</h3>


<p>
(<em>Appears on:</em><a href="#networkproblemdetector">NetworkProblemDetector</a>, <a href="#shootproviderconfig">ShootProviderConfig</a>)
</p>

<p>
AdditionalProbe defines a single network probe that is logically decoupled from the Shoot/Seed cluster topology.
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
<code>jobID</code></br>
<em>
string
</em>
</td>
<td>
<p>JobID is the unique identifier for this probe job.</p>
</td>
</tr>
<tr>
<td>
<code>protocol</code></br>
<em>
<a href="#probeprotocol">ProbeProtocol</a>
</em>
</td>
<td>
<p>Protocol is the probe protocol: TCP, HTTPS, or ICMP.</p>
</td>
</tr>
<tr>
<td>
<code>host</code></br>
<em>
string
</em>
</td>
<td>
<p>Host is the target hostname or IP address.<br />Required for all protocols. For TCP and HTTPS, used as the connection target.<br />For ICMP, used as the ping target.</p>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
integer
</em>
</td>
<td>
<p>Port is the target port (1-65535). Not used for ICMP probes.</p>
</td>
</tr>
<tr>
<td>
<code>period</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#duration-v1-meta">Duration</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Period optionally overrides the default check period for this probe.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="configuration">Configuration
</h3>


<p>
Configuration contains information about the network problem detector configuration.
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
<code>networkProblemDetector</code></br>
<em>
<a href="#networkproblemdetector">NetworkProblemDetector</a>
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
<a href="#healthcheckconfig">HealthCheckConfig</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HealthCheckConfig is the config for the health check controller.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="k8sexporter">K8sExporter
</h3>


<p>
(<em>Appears on:</em><a href="#networkproblemdetector">NetworkProblemDetector</a>)
</p>

<p>
K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.
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
boolean
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#duration-v1-meta">Duration</a>
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
float
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinFailingPeerNodeShare if > 0, reports node conditions `ClusterNetworkProblems` or `HostNetworkProblems` for node checks only if minimum share of destination peer nodes are failing. Valid range: [0.0,1.0]</p>
</td>
</tr>

</tbody>
</table>


<h3 id="networkproblemdetector">NetworkProblemDetector
</h3>


<p>
(<em>Appears on:</em><a href="#configuration">Configuration</a>)
</p>

<p>
NetworkProblemDetector contains the configuration for the network problem detector.
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#duration-v1-meta">Duration</a>
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
integer
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxPeerNodes optionally overrides the MaxPeerNodes in the agent config (maximum number of is the default period for jobs running in the agent.</p>
</td>
</tr>
<tr>
<td>
<code>icmpEnabled</code></br>
<em>
boolean
</em>
</td>
<td>
<em>(Optional)</em>
<p>IcmpEnabled is a flag if ICMP ping checks should be performed.</p>
</td>
</tr>
<tr>
<td>
<code>k8sExporter</code></br>
<em>
<a href="#k8sexporter">K8sExporter</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>K8sExporter configures the K8s exporter for updating node conditions and creating events.</p>
</td>
</tr>
<tr>
<td>
<code>additionalProbes</code></br>
<em>
<a href="#additionalprobe">AdditionalProbe</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalProbes defines additional probes that run independently of the Shoot/Seed cluster topology,<br />enabling infrastructure-level network diagnostics.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="probeprotocol">ProbeProtocol
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#additionalprobe">AdditionalProbe</a>)
</p>

<p>
ProbeProtocol defines the protocol for an additional probe.
</p>


<h3 id="shootproviderconfig">ShootProviderConfig
</h3>


<p>
ShootProviderConfig is the per-shoot configuration stored in Extension.spec.providerConfig.
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
<code>icmpEnabled</code></br>
<em>
boolean
</em>
</td>
<td>
<em>(Optional)</em>
<p>IcmpEnabled enables ICMP ping checks for this shoot.</p>
</td>
</tr>
<tr>
<td>
<code>additionalProbes</code></br>
<em>
<a href="#additionalprobe">AdditionalProbe</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalProbes defines additional probe jobs for this shoot.</p>
</td>
</tr>

</tbody>
</table>


