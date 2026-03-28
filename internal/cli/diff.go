package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/providers/incus"
	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff -f manifest.yaml",
		Short: "Show differences between manifest and actual state",
		Long: `Show differences between a manifest and the actual state of resources.

The diff command compares the desired state defined in a manifest with the
actual state of resources in the infrastructure. This is useful for previewing
changes before applying them.

Examples:
  # Show diff for a manifest
  omnigraph diff -f manifest.yaml

  # Show diff with detailed output
  omnigraph diff -f manifest.yaml --detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd)
		},
	}

	cmd.Flags().StringP("file", "f", "", "Path to manifest file (required)")
	cmd.Flags().Bool("detailed", false, "Show detailed differences")
	cmd.MarkFlagRequired("file")

	return cmd
}

func runDiff(cmd *cobra.Command) error {
	ctx := context.Background()

	manifestPath, _ := cmd.Flags().GetString("file")
	detailed, _ := cmd.Flags().GetBool("detailed")

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

	// Initialize Incus provider
	provider, err := incus.NewProvider(incus.Config{
		SocketPath: "/var/lib/incus/unix.socket",
	})
	if err != nil {
		return fmt.Errorf("failed to create Incus provider: %w", err)
	}

	fmt.Printf("Comparing manifest %s with actual state...\n\n", manifest.Metadata.Name)

	hasDifferences := false

	for _, resource := range manifest.Spec.Resources {
		exists, err := provider.Exists(ctx, resource)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to check %s/%s: %v\n",
				resource.Kind, resource.Metadata.Name, err)
			continue
		}

		if !exists {
			fmt.Printf("+ %s/%s (will be created)\n", resource.Kind, resource.Metadata.Name)
			if detailed {
				printResourceSpec(resource)
			}
			hasDifferences = true
			continue
		}

		// Get actual state
		actual, err := provider.GetActualState(ctx, resource)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to get actual state for %s/%s: %v\n",
				resource.Kind, resource.Metadata.Name, err)
			continue
		}

		// Compare specs
		if !specsEqual(resource.Spec, actual.Spec) {
			fmt.Printf("~ %s/%s (will be updated)\n", resource.Kind, resource.Metadata.Name)
			if detailed {
				fmt.Println("  Desired:")
				printResourceSpec(resource)
				fmt.Println("  Actual:")
				printResourceSpec(actual)
			}
			hasDifferences = true
		} else {
			fmt.Printf("= %s/%s (no changes)\n", resource.Kind, resource.Metadata.Name)
		}
	}

	if !hasDifferences {
		fmt.Println("\nNo differences found. All resources are in desired state.")
	} else {
		fmt.Println("\nDifferences found. Run 'omnigraph apply' to reconcile desired manifest state with the provider.")
	}

	return nil
}

func printResourceSpec(resource resources.Resource) {
	specYAML, _ := yaml.Marshal(resource.Spec)
	fmt.Printf("    %s\n", string(specYAML))
}
