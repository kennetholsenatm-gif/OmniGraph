package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/providers/incus"
	"github.com/kennetholsenatm-gif/omnigraph/internal/reconcile"
	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newApplyCmd() *cobra.Command {
	var manifestPath string
	var namespace string
	var dryRun bool
	var wait bool
	var timeout string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a declarative infrastructure manifest",
		Long: `Apply a declarative infrastructure manifest to create or update resources.

The apply command reads a YAML or JSON manifest file and reconciles the desired state
with the actual state of the infrastructure. This is similar to 'kubectl apply' but
for infrastructure resources like Incus containers, networks, and storage pools.

Examples:
  # Apply a manifest
  omnigraph apply -f manifest.yaml

  # Apply with dry-run (show diff only)
  omnigraph apply -f manifest.yaml --dry-run

  # Apply and wait for reconciliation
  omnigraph apply -f manifest.yaml --wait --timeout 10m

  # Apply to specific namespace
  omnigraph apply -f manifest.yaml --namespace production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd, manifestPath, namespace, dryRun, wait, timeout)
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "file", "f", "", "Path to manifest file (required)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to apply resources to")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Only show what would be changed, don't apply")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for reconciliation to complete")
	cmd.Flags().StringVar(&timeout, "timeout", "5m", "Timeout for wait (e.g., 30s, 5m, 1h)")
	cmd.MarkFlagRequired("file")

	return cmd
}

func runApply(cmd *cobra.Command, manifestPath, namespace string, dryRun, wait bool, timeoutStr string) error {
	ctx := context.Background()

	// Read manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var manifest resources.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Override namespace if provided
	if namespace != "" {
		manifest.Metadata.Namespace = namespace
	}

	// Validate manifest
	if err := validateManifest(&manifest); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Initialize reconciliation controller
	interval := 5 * time.Minute
	if manifest.Spec.Reconciliation != nil && manifest.Spec.Reconciliation.Interval != "" {
		parsed, err := time.ParseDuration(manifest.Spec.Reconciliation.Interval)
		if err == nil {
			interval = parsed
		}
	}

	controller := reconcile.NewController(interval)

	// Register Incus provider
	incusProvider, err := incus.NewProvider(incus.Config{
		SocketPath: "/var/lib/incus/unix.socket",
	})
	if err != nil {
		return fmt.Errorf("failed to create Incus provider: %w", err)
	}
	controller.RegisterProvider("incus", incusProvider)
	controller.RegisterProvider("lxd", incusProvider)
	controller.RegisterProvider("incusos", incusProvider)

	// Process each resource
	for i, resource := range manifest.Spec.Resources {
		fmt.Fprintf(cmd.OutOrStdout(), "Processing resource %d/%d: %s/%s\n",
			i+1, len(manifest.Spec.Resources), resource.Kind, resource.Metadata.Name)

		if dryRun {
			// Show what would be done
			if err := showDiff(ctx, controller, resource); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to compute diff for %s: %v\n",
					resource.Metadata.Name, err)
			}
			continue
		}

		// Add resource to controller
		if err := controller.AddResource(ctx, resource); err != nil {
			return fmt.Errorf("failed to add resource %s: %w", resource.Metadata.Name, err)
		}
	}

	if dryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "\nDry-run complete. No changes applied.")
		return nil
	}

	// Run reconciliation
	fmt.Fprintln(cmd.OutOrStdout(), "\nStarting reconciliation...")

	if wait {
		// Parse timeout
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}

		// Run with timeout
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := controller.Run(ctx); err != nil {
			if err == context.DeadlineExceeded {
				return fmt.Errorf("reconciliation timed out after %s", timeoutStr)
			}
			return fmt.Errorf("reconciliation failed: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Reconciliation completed successfully!")
	} else {
		// Start reconciliation in background
		go func() {
			if err := controller.Run(ctx); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Reconciliation error: %v\n", err)
			}
		}()

		fmt.Fprintln(cmd.OutOrStdout(), "Reconciliation started in background.")
		fmt.Fprintln(cmd.OutOrStdout(), "Use 'omnigraph get' to check resource status.")
	}

	return nil
}

func validateManifest(manifest *resources.Manifest) error {
	if manifest.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if manifest.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if manifest.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(manifest.Spec.Resources) == 0 {
		return fmt.Errorf("at least one resource is required")
	}

	// Validate each resource
	for i, resource := range manifest.Spec.Resources {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource %d: %w", i, err)
		}
	}

	return nil
}

func showDiff(ctx context.Context, controller *reconcile.Controller, resource resources.Resource) error {
	// Get provider for the resource
	providerName := resource.Spec.Provider
	if providerName == "" {
		providerName = "incus" // default
	}

	// Check if resource exists
	exists, err := controller.Provider(providerName).Exists(ctx, resource)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("  + %s/%s (will be created)\n", resource.Kind, resource.Metadata.Name)
		return nil
	}

	// Get actual state
	actual, err := controller.Provider(providerName).GetActualState(ctx, resource)
	if err != nil {
		return err
	}

	// Compare specs
	if !specsEqual(resource.Spec, actual.Spec) {
		fmt.Printf("  ~ %s/%s (will be updated)\n", resource.Kind, resource.Metadata.Name)
		// In a real implementation, show detailed diff
		return nil
	}

	fmt.Printf("  = %s/%s (no changes)\n", resource.Kind, resource.Metadata.Name)
	return nil
}

func specsEqual(a, b interface{}) bool {
	// Simple comparison - in production, use deep equal
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}