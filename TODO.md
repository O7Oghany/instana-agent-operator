## Next Steps

### ~~Fix K8Sensor Deployment And~~ Test Seamless Upgrade from deprecated k8s sensors

~~Deployment of K8Sensor is currently broken, fix this and then~~ run tests to ensure upgrade from a configuration
using the deprecated kubernetes sensor is seamless. To test this you should install the agent using the old (2.x)
version of the operator that uses helm and configure the agent to use the deprecated k8s sensor. Then upgrade the
operator to the new (3.x) version and ensure that the upgrade works seamlessly and that the new k8sensor is deployed.
Then check the Instana backend to ensure that k8s data collection has continued to function as expected through the
transition.

### Misconfiguration Errors

If the user misconfigures the agent then the attempt to create or update the Agent CR should be rejected. This can be
achieved using one of or a combination of the following methods.

#### CR Validation

[Validation rules](https://kubernetes.io/blog/2022/09/23/crd-validation-rules-beta/) and schema-based
[generation](https://book.kubebuilder.io/reference/markers/crd.html),
[validation](https://book.kubebuilder.io/reference/markers/crd-validation.html), and
[processing](https://book.kubebuilder.io/reference/markers/crd-processing.html) rules can be used to verify validity of
user-provided configuration and provide useful feedback for troubleshooting.

#### Webhook Validation

[Defaulting and Validation Webhooks](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation) could be used
for more advanced validation and to ensure defaulted values will appear on the CR present on the cluster without the
need for updates to the CR by the controller that could cause performance issues if another controller is managing the
agent CR.

#### Validation Admission Policy

Beginning in k8s v1.26 (alpha) or v1.28 (beta) a
[ValidationAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/) may
also be used to configure validation rules using [Common Expression Language (CEL)](https://github.com/google/cel-spec).

### Testing

#### Unit Tests

Unit test coverage exists now for most of the repo, but I recommend adding unit test coverage for
[status.go](./pkg/k8s/operator/status/status.go) and [event-filter.go](./controllers/event_filter.go). Unit tests should
also be added for the [k8sensor deployment builder](./pkg/k8s/object/builders/k8s-sensor/deployment/deployment.go). I
recommend implementing mostly whitebox testing. This should be very similar to the tests that already exist
[for the agent daemonset](./pkg/k8s/object/builders/agent/daemonset/daemonset_test.go). You will need to use the
[mock-gen](https://github.com/golang/mock) tool to generate mocks for some of the interfaces used by the deployment
builder. You will then need to add the relevant mock-gen commands to the Makefile, so they can be easily regenerated
in the future if needed. For an example of this, please see the mockgen commands for the daemonset tests
[in the Makefile](./Makefile#L227).

#### Behavioral Tests

Some behavioral tests exist at [controllers/suite_test.go](./controllers/suite_test.go). These will run a mock k8s
server and test the agent controller's behavior to ensure that the lifecycle of owned resources are managed as expected
during create, update, and delete of the Agent CR. Additional tests can be added here as needed or desired.

#### End to End Tests

The end to end tests that currently run in the helm and webhook builds will need to be setup to run against PRs for this
repo to ensure that the end-to-end functionality of the agent works as expected when changes are made to the operator
logic.

### Chart Update

The Helm chart should be updated to wrap the operator and an instance of the CR built by directly using toYaml on the
Values.yaml file to construct the spec. This should end up looking something like this:

```yaml
apiVersion: instana.io/v1
kind: InstanaAgent
metadata:
  name: "{{ .Release.Name }}"
  namespace: "{{ .Release.Namespace }}"
spec:
  {{- toYaml .Values | nindent 2}}
```

But you may need to modify it a bit to get the formatting exactly right. The new Helm chart should include only this
and the CRD, plus the resources needed to deploy the operator, which can be generated by running
`kustomize build config/default` and then replacing the namespace in each namespaced resource with
`{{ .Release.Namespace }}` to ensure that all resources will be deployed into the desired namespace. Helm itself will
ensure that the CRD will be installed before the InstanaAgent CR. See
[here](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/) for more information.

#### Chart Update Automation

Some automation to make updates to the chart based on changes to the CRD or
operator deployment may be desirable if these things are expected to change often.

## Future Considerations

### Manifest Build Automation

It may be useful to add logic to the PR builds that will automatically regenerate manifests and bundle YAMLs and
commit the changes to the source branch. Additionally, if you find that you are frequently making changes to operator
deployment files or sample CRs that would affect the Helm chart you may want to consider creating some automation that
will automatically push changes to the chart when relevant files in the operator repo have changed.

### Code Linting

Settings for [static code linting](.golangci.yml) should be reviewed and updated to suit the teams preferences. A basic
set of rules is currently enabled and will run during PR builds for this repo.

### Error Wrapping

Custom error types should be created with relevant messages to wrap errors that are being passed up the stack with
relevant information for debugging.

### Sensitive Data

Currently sensitive data (agent key, download, key, certificates, etc.) can be configured directly with the Agent CR.
This is considered bad-practice and can be a security risk in some cases; however, it may be alright to keep as a means
to deploy agents easily in development environments. Customers should be advised to place sensitive data directly into
secrets and reference the secrets from the agent spec.

### Configure Exponential Backoff

Rate-limiting [for the controller](https://danielmangum.com/posts/controller-runtime-client-go-rate-limiting/) should
be configured to prevent potential performance issues in cases where the cluster is inaccessible or the agent otherwise
cannot be deployed for some reason.

### Automatic Tolerations

Options could be added to the CR-spec to enable agents to run on master nodes by automatically setting the appropriate
tolerations for node taints.

### Automatic Zones

If desired an option could be added to automatically assign zone names to agent instances based on the value of the
`topology.kubernetes.io/zone` label on the node on which they are running.

### Logging

It may be worth considering the use of different default logging settings to improve readability
(e.g. --zap-time-encoding=rfc3339 --zap-encoder=console).

### .spec.agent.configuration_yaml

This could potentially be deprecated and replaced by a field using the `json.RawMessage` type, which would enable the
configuration yaml to be configured using native yaml within the CR rather than as an embedded string.

### Probes

The agent should have a readiness probe in addition to its liveness probe. The k8s_sensor should also have liveness and
readiness probes. A startup probe can also be added to the agent now that it is supported in all stable versions of k8s.
This will allow for faster readiness and recovery from soft-locks (if they occur) since it will allow the
initialDelaySeconds to be reduced on the liveness probe. The agent and k8sensor may also want to create dedicated
readiness endpoints to allow their actual availability to be reflected more accurately in the CR status. In the
agent's case, the availability of independent servers running on different ports may need to be considered when
deciding whether to do this since traffic directed at k8s services will not be forwarded to pods that are marked as
unready.

### PVs For Package Downloads

Optional Persistent volumes could potentially be used to cache dynamically downloaded updates and packages in between
agent restarts.

### Runtime Status

Runtime status information from the agent could be scraped and incorporated into the status tracked by the CR if this
is deemed useful.

### Ephemeral Storage Requests/Limits

Requests and limits for
[ephemeral storage](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#setting-requests-and-limits-for-local-ephemeral-storage)
can be set to ensure that agent pods are assigned to nodes containing appropriate storage space for dynamic agent
package downloads or to ensure agents do not exceed some limit of storage use on the host node.

### Certificate Generation

If desired, certificates could be automatically generated and configured when appropriate if cert-manager or
OpenShift's certificate generation is available.

### Network Policies

[Network policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) can be used to restrict
inbound traffic on ports that the agent or k8s_sensor do not use as a security measure. (May not work on agent itself
due to `hostNetwork: true).

### Auto-Reload on Agent-Key or Download-Key Change

Currently, the agent-key and download-key are read by the agent via environment variable set via referencing a key in
one or more k8s secrets. It would be beneficial to watch the secret(s) and trigger a restart of the agent daemsonset if
a change is detected in the secret(s).