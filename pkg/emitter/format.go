package emitter

// BackendFormat identifies an emitter target. All major IaC/packaging shapes are in scope;
// individual emitters are implemented incrementally.
const (
	OpenTofuHCL         = "opentofu-hcl"
	TerraformHCL        = "terraform-hcl"
	PulumiTypeScript    = "pulumi-typescript"
	PulumiPython        = "pulumi-python"
	PulumiGo            = "pulumi-go"
	AnsiblePlaybook     = "ansible-playbook"
	AnsibleInventoryINI = "ansible-inventory-ini"
	KubernetesYAML      = "kubernetes-yaml"
	HelmChart           = "helm-chart"
	PackerHCL           = "packer-hcl"
	DockerCompose       = "docker-compose"
	CloudFormationJSON  = "cloudformation-json"
	CloudFormationYAML  = "cloudformation-yaml"
	PuppetManifest      = "puppet-manifest"
	PuppetHiera         = "puppet-hiera"
)

// AllFormats returns every supported backend id (for CLI and docs parity).
func AllFormats() []string {
	return []string{
		OpenTofuHCL,
		TerraformHCL,
		PulumiTypeScript,
		PulumiPython,
		PulumiGo,
		AnsiblePlaybook,
		AnsibleInventoryINI,
		KubernetesYAML,
		HelmChart,
		PackerHCL,
		DockerCompose,
		CloudFormationJSON,
		CloudFormationYAML,
		PuppetManifest,
		PuppetHiera,
	}
}
