package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/enclave"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newEnclaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enclave",
		Short: "Manage Wasm enclaves and graph topologies",
		Long: `Enclave as Code (EaC) commands for managing Wasm enclaves.

The enclave command provides lifecycle management for Zero-Trust 
Execution Environments (ZTEEs) running QminiWasm-core workloads.

Examples:
  # Apply an enclave manifest
  omnigraph enclave apply -f edge-agent.yaml

  # Check enclave status
  omnigraph enclave status my-enclave

  # List all enclaves
  omnigraph enclave list

  # Synchronize graph topology
  omnigraph enclave graph-sync -f topology.yaml`,
	}

	cmd.AddCommand(newEnclaveApplyCmd())
	cmd.AddCommand(newEnclaveStatusCmd())
	cmd.AddCommand(newEnclaveListCmd())
	cmd.AddCommand(newEnclaveDeleteCmd())
	cmd.AddCommand(newEnclaveEnrollCmd())
	cmd.AddCommand(newEnclaveGraphSyncCmd())

	return cmd
}

func newEnclaveApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply -f <manifest>",
		Short: "Apply an enclave manifest",
		Long: `Apply an enclave manifest to create or update a Wasm enclave.

The manifest defines the enclave configuration including:
- Runtime settings (WasmEdge, memory limits)
- Trust boundary (ZTEE enrollment)
- Cognitive payload (ML models)
- Routing strategy (quantum fallback)`,
		RunE: runEnclaveApply,
	}

	cmd.Flags().StringP("file", "f", "", "Path to enclave manifest file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func runEnclaveApply(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Read manifest file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var manifest enclave.Enclave
	ext := filepath.Ext(filePath)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &manifest); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	// Create enclave manager
	manager := enclave.NewManager(".")

	// Validate manifest
	if err := manager.Validate(&manifest); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Deploy enclave
	ctx := context.Background()
	if err := manager.Deploy(ctx, &manifest); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Perform ZTEE enrollment if required
	if manifest.Spec.TrustBoundary.Enrollment == "strict" {
		fmt.Fprintf(cmd.OutOrStdout(), "Performing ZTEE enrollment...\n")
		if err := manager.Enroll(ctx, manifest.Metadata.Name); err != nil {
			return fmt.Errorf("enrollment failed: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Enclave '%s' deployed successfully\n", manifest.Metadata.Name)
	return nil
}

func newEnclaveStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Get enclave status",
		Long: `Get the current status of a deployed enclave including:
- Deployment phase
- ZTEE enrollment status
- Runtime metrics
- Health conditions`,
		Args: cobra.ExactArgs(1),
		RunE: runEnclaveStatus,
	}

	return cmd
}

func runEnclaveStatus(cmd *cobra.Command, args []string) error {
	enclaveName := args[0]

	manager := enclave.NewManager(".")
	status, err := manager.GetStatus(enclaveName)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Display status
	fmt.Fprintf(cmd.OutOrStdout(), "Enclave: %s\n", enclaveName)
	fmt.Fprintf(cmd.OutOrStdout(), "Phase: %s\n", status.Phase)

	if status.EnrollmentStatus != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Enrolled: %v\n", status.EnrollmentStatus.Enrolled)
		if !status.EnrollmentStatus.AttestedAt.IsZero() {
			fmt.Fprintf(cmd.OutOrStdout(), "Attested: %s\n", status.EnrollmentStatus.AttestedAt.Format(time.RFC3339))
		}
	}

	if status.RuntimeMetrics != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "\nRuntime Metrics:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Memory: %.2f MB\n", status.RuntimeMetrics.MemoryUsageMb)
		fmt.Fprintf(cmd.OutOrStdout(), "  CPU: %.2f%%\n", status.RuntimeMetrics.CPUPercent)
		fmt.Fprintf(cmd.OutOrStdout(), "  Inferences: %d\n", status.RuntimeMetrics.InferenceCount)
		fmt.Fprintf(cmd.OutOrStdout(), "  Avg Latency: %.2f ms\n", status.RuntimeMetrics.AvgLatencyMs)
	}

	if len(status.Conditions) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\nConditions:\n")
		for _, c := range status.Conditions {
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", c.Status, c.Type, c.Message)
		}
	}

	return nil
}

func newEnclaveListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all enclaves",
		Long: `List all deployed enclaves with their current status.`,
		RunE: runEnclaveList,
	}

	return cmd
}

func runEnclaveList(cmd *cobra.Command, args []string) error {
	manager := enclave.NewManager(".")
	enclaves := manager.List()

	if len(enclaves) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No enclaves deployed")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-20s %-15s\n", "NAME", "PHASE", "ENROLLED")
	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-20s %-15s\n", "----", "-----", "--------")

	for _, e := range enclaves {
		phase := "pending"
		enrolled := "no"

		if e.Status != nil {
			phase = e.Status.Phase
			if e.Status.EnrollmentStatus != nil && e.Status.EnrollmentStatus.Enrolled {
				enrolled = "yes"
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-20s %-15s\n", e.Metadata.Name, phase, enrolled)
	}

	return nil
}

func newEnclaveDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an enclave",
		Long: `Delete an enclave and clean up its resources.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runEnclaveDelete,
	}

	return cmd
}

func runEnclaveDelete(cmd *cobra.Command, args []string) error {
	enclaveName := args[0]

	manager := enclave.NewManager(".")
	if err := manager.Delete(enclaveName); err != nil {
		return fmt.Errorf("failed to delete enclave: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Enclave '%s' deleted successfully\n", enclaveName)
	return nil
}

func newEnclaveEnrollCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enroll <name>",
		Short: "Enroll an enclave with ZTEE",
		Long: `Perform Zero-Trust Enclave Enrollment (ZTEE) for an enclave.

This establishes cryptographic trust between the enclave and the
attestation provider (Keycloak, Vault, etc.).`,
		Args: cobra.ExactArgs(1),
		RunE: runEnclaveEnroll,
	}

	cmd.Flags().String("provider", "keycloak", "Attestation provider (keycloak, vault, custom)")

	return cmd
}

func runEnclaveEnroll(cmd *cobra.Command, args []string) error {
	enclaveName := args[0]
	provider, _ := cmd.Flags().GetString("provider")

	manager := enclave.NewManager(".")
	ctx := context.Background()

	fmt.Fprintf(cmd.OutOrStdout(), "Enrolling enclave '%s' with %s...\n", enclaveName, provider)

	if err := manager.Enroll(ctx, enclaveName); err != nil {
		return fmt.Errorf("enrollment failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Enclave '%s' enrolled successfully\n", enclaveName)
	return nil
}

func newEnclaveGraphSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph-sync -f <topology>",
		Short: "Synchronize graph topology",
		Long: `Synchronize an enclave graph topology with the graph connector.

The topology file defines the relationships between enclaves,
including communication channels, dependencies, and routing.`,
		RunE: runGraphSync,
	}

	cmd.Flags().StringP("file", "f", "", "Path to graph topology file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func runGraphSync(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Read topology file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read topology: %w", err)
	}

	// Parse topology
	var graph enclave.EnclaveGraph
	if err := yaml.Unmarshal(data, &graph); err != nil {
		return fmt.Errorf("failed to parse topology: %w", err)
	}

	// Create graph connector
	manager := enclave.NewManager(".")
	connector := enclave.NewGraphConnector(manager)

	ctx := context.Background()
	if err := connector.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer connector.Disconnect()

	// Sync graph
	if err := connector.SyncGraph(ctx, &graph); err != nil {
		return fmt.Errorf("graph sync failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Graph topology synced successfully\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Nodes: %d\n", len(graph.Nodes))
	fmt.Fprintf(cmd.OutOrStdout(), "  Edges: %d\n", len(graph.Edges))

	return nil
}