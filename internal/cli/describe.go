package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/providers/incus"
	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [resource-type] [name]",
		Short: "Describe a resource",
		Long: `Describe a resource in detail.

The describe command shows detailed information about a resource including
its configuration, status, and conditions.

Examples:
  # Describe an instance
  omnigraph describe instance web-1

  # Describe a network
  omnigraph describe network web-network

  # Describe a storage pool
  omnigraph describe pool production-storage

  # Describe a profile
  omnigraph describe profile web-server`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribe(cmd, args)
		},
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize Incus provider
	provider, err := incus.NewProvider(incus.Config{
		SocketPath: "/var/lib/incus/unix.socket",
	})
	if err != nil {
		return fmt.Errorf("failed to create Incus provider: %w", err)
	}

	resourceType := args[0]
	resourceName := args[1]

	resource := &resources.Resource{
		Kind: getKindFromType(resourceType),
		Metadata: resources.ResourceMetadata{
			Name: resourceName,
		},
	}

	switch resourceType {
	case "instance", "container", "vm":
		return describeInstance(ctx, provider, resource)
	case "network":
		return describeNetwork(ctx, provider, resource)
	case "pool", "storage":
		return describeStoragePool(ctx, provider, resource)
	case "profile":
		return describeProfile(ctx, provider, resource)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

func describeInstance(ctx context.Context, provider *incus.Provider, resource *resources.Resource) error {
	instance, err := provider.GetInstance(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	fmt.Printf("Name: %s\n", instance.Name)
	fmt.Printf("Type: %s\n", instance.Type)
	fmt.Printf("Status: %s\n", instance.Status)
	fmt.Printf("Architecture: %s\n", instance.Architecture)
	fmt.Printf("Location: %s\n", instance.Location)
	fmt.Printf("Project: %s\n", instance.Project)
	fmt.Printf("Created: %s\n", instance.CreatedAt)
	fmt.Printf("Last Used: %s\n", instance.LastUsedAt)
	fmt.Printf("Ephemeral: %v\n", instance.Ephemeral)
	fmt.Printf("Profiles: %v\n", instance.Profiles)

	fmt.Printf("\nNetwork:\n")
	fmt.Printf("  IPv4: %s\n", formatIPs(instance.IPv4))
	fmt.Printf("  IPv6: %s\n", formatIPs(instance.IPv6))

	fmt.Printf("\nConfiguration:\n")
	for k, v := range instance.Config {
		fmt.Printf("  %s: %s\n", k, v)
	}

	fmt.Printf("\nDevices:\n")
	for name, dev := range instance.Devices {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    type: %s\n", dev.Type)
		if dev.Name != "" {
			fmt.Printf("    name: %s\n", dev.Name)
		}
		if dev.Parent != "" {
			fmt.Printf("    parent: %s\n", dev.Parent)
		}
		if dev.NICType != "" {
			fmt.Printf("    nictype: %s\n", dev.NICType)
		}
		if dev.Path != "" {
			fmt.Printf("    path: %s\n", dev.Path)
		}
		if dev.Pool != "" {
			fmt.Printf("    pool: %s\n", dev.Pool)
		}
		if dev.Size != "" {
			fmt.Printf("    size: %s\n", dev.Size)
		}
		for k, v := range dev.Properties {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	fmt.Printf("\nState:\n")
	stateJSON, _ := json.MarshalIndent(instance, "  ", "  ")
	fmt.Printf("  %s\n", string(stateJSON))

	return nil
}

func describeNetwork(ctx context.Context, provider *incus.Provider, resource *resources.Resource) error {
	network, err := provider.GetNetwork(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	fmt.Printf("Name: %s\n", network.Name)
	fmt.Printf("Type: %s\n", network.Type)
	fmt.Printf("Managed: %v\n", network.Managed)
	fmt.Printf("Description: %s\n", network.Description)

	fmt.Printf("\nConfiguration:\n")
	for k, v := range network.Config {
		fmt.Printf("  %s: %s\n", k, v)
	}

	fmt.Printf("\nUsed By:\n")
	for _, u := range network.UsedBy {
		fmt.Printf("  %s\n", u)
	}

	fmt.Printf("\nState:\n")
	stateJSON, _ := json.MarshalIndent(network, "  ", "  ")
	fmt.Printf("  %s\n", string(stateJSON))

	return nil
}

func describeStoragePool(ctx context.Context, provider *incus.Provider, resource *resources.Resource) error {
	pool, err := provider.GetStoragePool(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get storage pool: %w", err)
	}

	fmt.Printf("Name: %s\n", pool.Name)
	fmt.Printf("Driver: %s\n", pool.Driver)
	fmt.Printf("Description: %s\n", pool.Description)

	fmt.Printf("\nConfiguration:\n")
	for k, v := range pool.Config {
		fmt.Printf("  %s: %s\n", k, v)
	}

	fmt.Printf("\nUsed By:\n")
	for _, u := range pool.UsedBy {
		fmt.Printf("  %s\n", u)
	}

	fmt.Printf("\nState:\n")
	stateJSON, _ := json.MarshalIndent(pool, "  ", "  ")
	fmt.Printf("  %s\n", string(stateJSON))

	return nil
}

func describeProfile(ctx context.Context, provider *incus.Provider, resource *resources.Resource) error {
	profile, err := provider.GetProfile(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	fmt.Printf("Name: %s\n", profile.Name)
	fmt.Printf("Description: %s\n", profile.Description)

	fmt.Printf("\nConfiguration:\n")
	for k, v := range profile.Config {
		fmt.Printf("  %s: %s\n", k, v)
	}

	fmt.Printf("\nDevices:\n")
	for name, dev := range profile.Devices {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    type: %s\n", dev.Type)
		if dev.Name != "" {
			fmt.Printf("    name: %s\n", dev.Name)
		}
		if dev.Parent != "" {
			fmt.Printf("    parent: %s\n", dev.Parent)
		}
		if dev.NICType != "" {
			fmt.Printf("    nictype: %s\n", dev.NICType)
		}
		if dev.Path != "" {
			fmt.Printf("    path: %s\n", dev.Path)
		}
		if dev.Pool != "" {
			fmt.Printf("    pool: %s\n", dev.Pool)
		}
		if dev.Size != "" {
			fmt.Printf("    size: %s\n", dev.Size)
		}
		for k, v := range dev.Properties {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	fmt.Printf("\nUsed By:\n")
	for _, u := range profile.UsedBy {
		fmt.Printf("  %s\n", u)
	}

	fmt.Printf("\nState:\n")
	stateJSON, _ := json.MarshalIndent(profile, "  ", "  ")
	fmt.Printf("  %s\n", string(stateJSON))

	return nil
}