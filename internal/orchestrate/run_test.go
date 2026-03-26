package orchestrate

import (
	"context"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
)

type fakeRunner struct {
	planJSON []byte
}

func (f *fakeRunner) Run(ctx context.Context, s runner.Step) (*runner.Result, error) {
	joined := strings.Join(s.Argv, " ")
	if strings.Contains(joined, "show") && strings.Contains(joined, "-json") {
		return &runner.Result{ExitCode: 0, Stdout: append([]byte(nil), f.planJSON...)}, nil
	}
	return &runner.Result{ExitCode: 0}, nil
}

func TestRun_SkipAnsibleWithFakeRunner(t *testing.T) {
	dir := t.TempDir()
	schema := filepath.Join(dir, ".omnigraph.schema")
	sample := []byte(`apiVersion: omnigraph/v1alpha1
kind: Project
metadata:
  name: t
spec: {}
`)
	if err := os.WriteFile(schema, sample, 0o600); err != nil {
		t.Fatal(err)
	}
	state := []byte(`{"version":4,"values":{"outputs":{},"root_module":{"resources":[]}}}`)
	if err := os.WriteFile(filepath.Join(dir, "terraform.tfstate"), state, 0o600); err != nil {
		t.Fatal(err)
	}
	planJSON := []byte(`{"planned_values":{"outputs":{},"root_module":{"resources":[]}}}`)
	graphOut := filepath.Join(dir, "graph.json")
	o := Options{
		Workdir:     dir,
		SchemaPath:  ".omnigraph.schema",
		AutoApprove: true,
		SkipAnsible: true,
		GraphOut:    graphOut,
		Runner:      "exec",
	}
	fr := &fakeRunner{planJSON: planJSON}
	if err := Run(context.Background(), fr, o, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(graphOut); err != nil {
		t.Fatal(err)
	}
}

func TestOptions_step_Container(t *testing.T) {
	o := Options{Runner: "container", TofuImage: "img:t"}
	s := o.step("tofu-plan", []string{"tofu", "plan"}, nil, "/tmp/work", "")
	if s.ContainerImage != "img:t" {
		t.Fatalf("image %q", s.ContainerImage)
	}
	if len(s.Mounts) != 1 || s.Mounts[0].HostPath != "/tmp/work" {
		t.Fatalf("mounts %+v", s.Mounts)
	}
}

func TestOptions_step_Container_AnsibleDualMount(t *testing.T) {
	o := Options{Runner: "container", AnsibleImage: "img:ansible"}
	s := o.step("ansible-check", []string{"ansible-playbook", "x.yml"}, nil, "/tmp/tofu", "/tmp/ansible")
	if s.ContainerImage != "img:ansible" {
		t.Fatalf("image %q", s.ContainerImage)
	}
	if len(s.Mounts) != 2 {
		t.Fatalf("want 2 mounts, got %+v", s.Mounts)
	}
	if s.Mounts[0].HostPath != "/tmp/tofu" || s.Mounts[0].ContainerPath != "/workspace" {
		t.Fatalf("mount0 %+v", s.Mounts[0])
	}
	if s.Mounts[1].HostPath != "/tmp/ansible" || s.Mounts[1].ContainerPath != "/ansible" {
		t.Fatalf("mount1 %+v", s.Mounts[1])
	}
}

func TestContainerAnsiblePlaybookPath(t *testing.T) {
	ansible := filepath.Join(t.TempDir(), "ansible")
	tofu := filepath.Join(t.TempDir(), "tofu")
	play := filepath.Join(ansible, "roles", "site.yml")
	p := containerAnsiblePlaybookPath(tofu, ansible, play, "roles/site.yml", "container")
	want := pathpkg.Join("/ansible", "roles/site.yml")
	if p != want {
		t.Fatalf("got %q want %q", p, want)
	}
	pb := filepath.Join(tofu, "playbooks", "x.yml")
	p = containerAnsiblePlaybookPath(tofu, "", pb, pb, "container")
	want2 := pathpkg.Join("/workspace", "playbooks/x.yml")
	if p != want2 {
		t.Fatalf("got %q want %q", p, want2)
	}
}
