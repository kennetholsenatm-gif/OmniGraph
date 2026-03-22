package runner

import (
	"strings"
	"testing"
)

func TestBuildContainerRunArgs(t *testing.T) {
	s := Step{
		ContainerImage:   "ghcr.io/opentofu/opentofu:1.8",
		ContainerWorkdir: "/workspace",
		Mounts: []VolumeMount{
			{HostPath: "/host/proj", ContainerPath: "/workspace", ReadOnly: false},
		},
		Env: map[string]string{"TF_VAR_project_name": "demo"},
		Argv: []string{"tofu", "version"},
	}
	args, err := BuildContainerRunArgs("docker", s)
	if err != nil {
		t.Fatal(err)
	}
	if args[0] != "docker" || args[1] != "run" {
		t.Fatalf("args: %v", args)
	}
	got := strings.Join(args, " ")
	if !strings.Contains(got, "-v /host/proj:/workspace:rw") {
		t.Fatalf("missing mount: %s", got)
	}
	if !strings.Contains(got, "-e TF_VAR_project_name=demo") {
		t.Fatalf("missing env: %s", got)
	}
	if !strings.Contains(got, "ghcr.io/opentofu/opentofu:1.8") {
		t.Fatal(got)
	}
	if !strings.Contains(got, "tofu version") {
		t.Fatal(got)
	}
}

func TestBuildContainerRunArgs_RejectsEmptyMount(t *testing.T) {
	_, err := BuildContainerRunArgs("docker", Step{
		ContainerImage: "x",
		Mounts:         []VolumeMount{{HostPath: "", ContainerPath: "/w"}},
		Argv:           []string{"sh"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
