package orchestrate

import (
	"bufio"
	"context"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
	"github.com/kennetholsenatm-gif/omnigraph/internal/plan"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"golang.org/x/term"
)

// Default images (override via Options).
const (
	DefaultTofuImage    = "ghcr.io/opentofu/opentofu:1.8"
	DefaultAnsibleImage = "cytopia/ansible:latest"
	DefaultPulumiImage  = "pulumi/pulumi:latest"
)

// Options configures the magic-handoff pipeline (plan → check → approve → apply → Ansible).
type Options struct {
	Workdir   string
	SchemaPath string // default .omnigraph.schema; joined with Workdir if not absolute
	Playbook  string // relative to Workdir, or to AnsibleRoot if set, or absolute under that root (required unless SkipAnsible)
	// AnsibleRoot is an optional second checkout; when set, Playbook is resolved under this directory and container runner mounts it at /ansible.
	AnsibleRoot string
	TFBinary  string // tofu or terraform (exec PATH; container argv[0])
	PlanFile  string // relative to Workdir, default tfplan
	StateFile string // relative to Workdir, default terraform.tfstate

	Runner           string // exec (default) or container
	ContainerRuntime string // docker or podman; empty = runner.DetectContainerRuntime()
	TofuImage        string
	AnsibleImage     string
	// PulumiImage is used when --iac-engine=pulumi is wired to container steps (default: DefaultPulumiImage).
	PulumiImage string

	AutoApprove bool
	GraphOut    string
	// TelemetryFile merges omnigraph/telemetry/v1 nodes/edges into graph output when GraphOut is set.
	TelemetryFile string

	// SkipAnsible skips ansible-playbook steps (tests / tofu-only workspaces).
	SkipAnsible bool

	// IACEngine selects the infrastructure CLI family (only tofu is implemented).
	IACEngine string // tofu (default) or pulumi (stub)
}

// Run executes validate → coerce → tofu plan → show json → inventory + ansible check → approve → apply → ansible apply.
func Run(ctx context.Context, r runner.Runner, o Options, log func(phase, detail string)) error {
	if log == nil {
		log = func(_, _ string) {}
	}
	engine := strings.ToLower(strings.TrimSpace(o.IACEngine))
	if engine == "" {
		engine = "tofu"
	}
	switch engine {
	case "tofu":
		// OpenTofu/Terraform argv path below.
	case "pulumi":
		return fmt.Errorf("orchestrate: --iac-engine=pulumi is not implemented yet; use tofu, or run Pulumi via ContainerRunner (see docs/execution-matrix.md); default image when wired: %s", o.pulumiImage())
	default:
		return fmt.Errorf("orchestrate: unknown --iac-engine=%q (supported: tofu)", engine)
	}
	if o.Workdir == "" {
		return fmt.Errorf("orchestrate: Workdir is required")
	}
	workAbs, err := filepath.Abs(o.Workdir)
	if err != nil {
		return fmt.Errorf("orchestrate: workdir: %w", err)
	}
	if o.Playbook == "" && !o.SkipAnsible {
		return fmt.Errorf("orchestrate: Playbook is required (or set SkipAnsible)")
	}
	ansibleRootArg := strings.TrimSpace(o.AnsibleRoot)
	var ansibleAbs string
	if ansibleRootArg != "" {
		ansibleAbs, err = filepath.Abs(ansibleRootArg)
		if err != nil {
			return fmt.Errorf("orchestrate: ansible-root: %w", err)
		}
	}
	schemaPath := o.SchemaPath
	if schemaPath == "" {
		schemaPath = ".omnigraph.schema"
	}
	if !filepath.IsAbs(schemaPath) {
		schemaPath = filepath.Join(workAbs, schemaPath)
	}
	playRel := strings.TrimSpace(o.Playbook)
	var playHost string
	if !o.SkipAnsible {
		if ansibleAbs != "" {
			if filepath.IsAbs(playRel) {
				playHost = filepath.Clean(playRel)
				relAP, err := filepath.Rel(ansibleAbs, playHost)
				if err != nil || strings.HasPrefix(relAP, "..") {
					return fmt.Errorf("orchestrate: playbook %q must be under --ansible-root %q", playRel, ansibleRootArg)
				}
			} else {
				playHost = filepath.Join(ansibleAbs, filepath.FromSlash(playRel))
			}
		} else {
			playHost = playRel
			if !filepath.IsAbs(playHost) {
				playHost = filepath.Join(workAbs, playRel)
			}
		}
	}
	planRel := o.PlanFile
	if planRel == "" {
		planRel = "tfplan"
	}
	stateRel := o.StateFile
	if stateRel == "" {
		stateRel = "terraform.tfstate"
	}
	stateHost := filepath.Join(workAbs, stateRel)
	planJSONHost := filepath.Join(workAbs, ".omnigraph-plan.json")

	tfBin := o.TFBinary
	if tfBin == "" {
		tfBin = "tofu"
	}

	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("orchestrate: read schema: %w", err)
	}
	log("validate", schemaPath)
	if _, err := schema.ValidateRawDocument(raw); err != nil {
		return fmt.Errorf("orchestrate: validate: %w", err)
	}
	doc, err := project.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("orchestrate: parse document: %w", err)
	}
	art, err := coerce.FromDocument(doc)
	if err != nil {
		return fmt.Errorf("orchestrate: coerce: %w", err)
	}
	env := MergeExecutionEnv(art)

	runnerKind := strings.ToLower(strings.TrimSpace(o.Runner))
	if runnerKind == "" {
		runnerKind = "exec"
	}

	// --- tofu plan ---
	planArgv := []string{tfBin, "plan", "-out", hostOrContainerPath(runnerKind, planRel)}
	sPlan := o.step("tofu-plan", planArgv, env, workAbs, "")
	log("plan", strings.Join(planArgv, " "))
	res, err := r.Run(ctx, sPlan)
	if err != nil {
		return fmt.Errorf("orchestrate: plan run: %w", err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("orchestrate: tofu plan failed (%d): %s", res.ExitCode, string(res.Stderr))
	}

	// --- tofu show -json ---
	showArgv := []string{tfBin, "show", "-json", hostOrContainerPath(runnerKind, planRel)}
	sShow := o.step("tofu-show", showArgv, env, workAbs, "")
	log("plan-json", strings.Join(showArgv, " "))
	res, err = r.Run(ctx, sShow)
	if err != nil {
		return fmt.Errorf("orchestrate: show: %w", err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("orchestrate: tofu show failed (%d): %s", res.ExitCode, string(res.Stderr))
	}
	if err := os.WriteFile(planJSONHost, res.Stdout, 0o600); err != nil {
		return fmt.Errorf("orchestrate: write plan json: %w", err)
	}
	defer func() { _ = os.Remove(planJSONHost) }()

	pj, err := plan.Parse(res.Stdout)
	if err != nil {
		return fmt.Errorf("orchestrate: parse plan json: %w", err)
	}
	projHosts := plan.ProjectedHosts(pj)
	invCheck := filepath.Join(workAbs, fmt.Sprintf(".omnigraph-check-%s.ini", sanitize(planRel)))
	invCheckContent := inventory.BuildINI(projHosts)
	if err := os.WriteFile(invCheck, []byte(invCheckContent), 0o600); err != nil {
		return fmt.Errorf("orchestrate: write check inventory: %w", err)
	}
	defer func() { _ = os.Remove(invCheck) }()

	if !o.SkipAnsible {
		invArg := invCheck
		playArg := playHost
		av := []string{"ansible-playbook", "--check", "-i", invArg, playArg}
		if runnerKind == "container" {
			invArg = pathpkg.Join("/workspace", filepath.Base(invCheck))
			playArg = containerAnsiblePlaybookPath(workAbs, ansibleAbs, playHost, playRel, runnerKind)
			av = []string{"ansible-playbook", "--check", "-i", invArg, playArg}
		}
		sAns := o.step("ansible-check", av, env, workAbs, ansibleAbs)
		log("ansible-check", strings.Join(av, " "))
		res, err = r.Run(ctx, sAns)
		if err != nil {
			return fmt.Errorf("orchestrate: ansible check: %w", err)
		}
		if res.ExitCode != 0 {
			return fmt.Errorf("orchestrate: ansible-playbook --check failed (%d): %s", res.ExitCode, string(res.Stderr))
		}
	}

	if err := confirmApply(o.AutoApprove); err != nil {
		return err
	}

	applyArgv := []string{tfBin, "apply", "-auto-approve", planRel}
	if runnerKind == "container" {
		applyArgv = []string{tfBin, "apply", "-auto-approve", pathpkg.Join("/workspace", filepath.ToSlash(planRel))}
	}
	sApply := o.step("tofu-apply", applyArgv, env, workAbs, "")
	log("apply", strings.Join(applyArgv, " "))
	res, err = r.Run(ctx, sApply)
	if err != nil {
		return fmt.Errorf("orchestrate: apply run: %w", err)
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("orchestrate: tofu apply failed (%d): %s", res.ExitCode, string(res.Stderr))
	}

	st, err := state.Load(stateHost)
	if err != nil {
		return fmt.Errorf("orchestrate: load state: %w", err)
	}
	liveHosts := state.ExtractHosts(st)
	invApply := filepath.Join(workAbs, fmt.Sprintf(".omnigraph-apply-%s.ini", sanitize(planRel)))
	if err := os.WriteFile(invApply, []byte(inventory.BuildINI(liveHosts)), 0o600); err != nil {
		return fmt.Errorf("orchestrate: write apply inventory: %w", err)
	}
	defer func() { _ = os.Remove(invApply) }()

	if !o.SkipAnsible {
		invArg := invApply
		playArg := playHost
		av := []string{"ansible-playbook", "-i", invArg, playArg}
		if runnerKind == "container" {
			invArg = pathpkg.Join("/workspace", filepath.Base(invApply))
			playArg = containerAnsiblePlaybookPath(workAbs, ansibleAbs, playHost, playRel, runnerKind)
			av = []string{"ansible-playbook", "-i", invArg, playArg}
		}
		sAns := o.step("ansible-apply", av, env, workAbs, ansibleAbs)
		log("ansible-apply", strings.Join(av, " "))
		res, err = r.Run(ctx, sAns)
		if err != nil {
			return fmt.Errorf("orchestrate: ansible apply: %w", err)
		}
		if res.ExitCode != 0 {
			return fmt.Errorf("orchestrate: ansible-playbook failed (%d): %s", res.ExitCode, string(res.Stderr))
		}
	}

	if o.GraphOut != "" {
		telemetryPath := o.TelemetryFile
		if telemetryPath != "" && !filepath.IsAbs(telemetryPath) {
			telemetryPath = filepath.Join(workAbs, telemetryPath)
		}
		gopts := graph.EmitOptions{
			PlanJSONPath:   planJSONHost,
			TerraformState: st,
			TelemetryPath:  telemetryPath,
		}
		gdoc, err := graph.Emit(doc, art, gopts)
		if err != nil {
			return fmt.Errorf("orchestrate: graph emit: %w", err)
		}
		gdoc.Spec.Phase = "apply"
		for i := range gdoc.Spec.Phases {
			if gdoc.Spec.Phases[i].Name == "plan" {
				gdoc.Spec.Phases[i].Status = "ok"
			}
			if gdoc.Spec.Phases[i].Name == "apply" {
				gdoc.Spec.Phases[i].Status = "ok"
			}
		}
		b, err := graph.EncodeIndent(gdoc)
		if err != nil {
			return err
		}
		if err := os.WriteFile(o.GraphOut, b, 0o644); err != nil {
			return fmt.Errorf("orchestrate: graph out: %w", err)
		}
		log("graph", o.GraphOut)
	}

	return nil
}

// containerAnsiblePlaybookPath returns the playbook path to pass inside the Ansible container.
func containerAnsiblePlaybookPath(workAbs, ansibleAbs, playHost, playRel, runnerKind string) string {
	if runnerKind != "container" {
		return playHost
	}
	if ansibleAbs != "" {
		rel, err := filepath.Rel(ansibleAbs, playHost)
		if err != nil {
			return pathpkg.Join("/ansible", filepath.ToSlash(playRel))
		}
		return pathpkg.Join("/ansible", filepath.ToSlash(rel))
	}
	if filepath.IsAbs(playRel) {
		rel, err := filepath.Rel(workAbs, playHost)
		if err != nil || strings.HasPrefix(rel, "..") {
			return pathpkg.Join("/workspace", filepath.ToSlash(playRel))
		}
		return pathpkg.Join("/workspace", filepath.ToSlash(rel))
	}
	return pathpkg.Join("/workspace", filepath.ToSlash(playRel))
}

func hostOrContainerPath(runnerKind, rel string) string {
	rel = filepath.ToSlash(rel)
	if runnerKind == "container" {
		return pathpkg.Join("/workspace", rel)
	}
	return rel
}

func sanitize(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '_')
		}
	}
	if len(b) == 0 {
		return "x"
	}
	return string(b)
}

func confirmApply(auto bool) error {
	if auto {
		return nil
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("orchestrate: apply requires --auto-approve when stdin is not a terminal")
	}
	fmt.Fprint(os.Stderr, "Apply infrastructure? [y/N]: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return fmt.Errorf("orchestrate: read confirmation: %w", err)
	}
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "y" || line == "yes" {
		return nil
	}
	return fmt.Errorf("orchestrate: apply cancelled")
}

func (o *Options) step(name string, argv []string, env map[string]string, workAbs, ansibleHostAbs string) runner.Step {
	kind := strings.ToLower(strings.TrimSpace(o.Runner))
	if kind == "" {
		kind = "exec"
	}
	s := runner.Step{Name: name, Argv: argv, Env: env}
	if kind != "container" {
		s.Dir = workAbs
		return s
	}
	rt := o.ContainerRuntime
	if rt == "" {
		rt = runner.DetectContainerRuntime()
	}
	if rt == "" {
		rt = "docker"
	}
	img := o.tofuImage()
	if strings.HasPrefix(name, "ansible") {
		img = o.ansibleImage()
	}
	s.ContainerImage = img
	s.ContainerWorkdir = "/workspace"
	mounts := []runner.VolumeMount{{HostPath: workAbs, ContainerPath: "/workspace", ReadOnly: false}}
	if ansibleHostAbs != "" && strings.HasPrefix(name, "ansible") {
		mounts = append(mounts, runner.VolumeMount{HostPath: ansibleHostAbs, ContainerPath: "/ansible", ReadOnly: false})
	}
	s.Mounts = mounts
	return s
}

func (o *Options) tofuImage() string {
	if o.TofuImage != "" {
		return o.TofuImage
	}
	return DefaultTofuImage
}

func (o *Options) ansibleImage() string {
	if o.AnsibleImage != "" {
		return o.AnsibleImage
	}
	return DefaultAnsibleImage
}

func (o *Options) pulumiImage() string {
	if o.PulumiImage != "" {
		return o.PulumiImage
	}
	return DefaultPulumiImage
}
