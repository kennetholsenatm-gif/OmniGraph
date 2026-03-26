package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kennetholsenatm-gif/omnigraph/internal/providers/incus"
	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var namespace string
	var output string

	cmd := &cobra.Command{
		Use:   "get [resource-type] [name]",
		Short: "Get resources",
		Long: `Get one or more resources.

The get command lists resources of a specific type or gets details of a single resource.

Examples:
  # List all instances
  omnigraph get instances

  # List all networks
  omnigraph get networks

  # Get specific instance
  omnigraph get instance web-1

  # Get in JSON format
  omnigraph get instances -o json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, args, namespace, output)
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter by namespace")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func runGet(cmd *cobra.Command, args []string, namespace, output string) error {
	ctx := context.Background()

	// Initialize Incus provider
	provider, err := incus.NewProvider(incus.Config{
		SocketPath: "/var/lib/incus/unix.socket",
	})
	if err != nil {
		return fmt.Errorf("failed to create Incus provider: %w", err)
	}

	resourceType := args[0]
	var resource *resources.Resource
	if len(args) > 1 {
		resource = &resources.Resource{
			Kind: getKindFromType(resourceType),
			Metadata: resources.ResourceMetadata{
				Name: args[1],
			},
		}
	}

	switch resourceType {
	case "instance", "instances", "container", "containers", "vm", "vms":
		if resource != nil {
			return getInstance(ctx, provider, resource, output)
		}
		return listInstances(ctx, provider, namespace, output)
	case "network", "networks":
		if resource != nil {
			return getNetwork(ctx, provider, resource, output)
		}
		return listNetworks(ctx, provider, namespace, output)
	case "pool", "pools", "storage", "storage-pools":
		if resource != nil {
			return getStoragePool(ctx, provider, resource, output)
		}
		return listStoragePools(ctx, provider, namespace, output)
	case "profile", "profiles":
		if resource != nil {
			return getProfile(ctx, provider, resource, output)
		}
		return listProfiles(ctx, provider, namespace, output)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

func getKindFromType(resourceType string) string {
	switch resourceType {
	case "instance", "instances", "container", "containers", "vm", "vms":
		return "ComputeInstance"
	case "network", "networks":
		return "Network"
	case "pool", "pools", "storage", "storage-pools":
		return "StoragePool"
	case "profile", "profiles":
		return "Profile"
	default:
		return resourceType
	}
}

func listInstances(ctx context.Context, provider *incus.Provider, namespace, output string) error {
	instances, err := provider.ListInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(instances)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSTATE\tIPV4\tIPV6\tARCHITECTURE")
	for _, inst := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			inst.Name,
			inst.Type,
			inst.Status,
			formatIPs(inst.IPv4),
			formatIPs(inst.IPv6),
			inst.Architecture,
		)
	}
	return w.Flush()
}

func getInstance(ctx context.Context, provider *incus.Provider, resource *resources.Resource, output string) error {
	instance, err := provider.GetInstance(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(instance)
	}

	fmt.Printf("Name: %s\n", instance.Name)
	fmt.Printf("Type: %s\n", instance.Type)
	fmt.Printf("State: %s\n", instance.Status)
	fmt.Printf("Architecture: %s\n", instance.Architecture)
	fmt.Printf("IPv4: %s\n", formatIPs(instance.IPv4))
	fmt.Printf("IPv6: %s\n", formatIPs(instance.IPv6))
	fmt.Printf("Created: %s\n", instance.CreatedAt)
	fmt.Printf("Last Used: %s\n", instance.LastUsedAt)
	fmt.Printf("Profiles: %v\n", instance.Profiles)
	fmt.Printf("Project: %s\n", instance.Project)

	return nil
}

func listNetworks(ctx context.Context, provider *incus.Provider, namespace, output string) error {
	networks, err := provider.ListNetworks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(networks)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tMANAGED\tIPV4\tIPV6\tDESCRIPTION")
	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%v\t%s\t%s\t%s\n",
			net.Name,
			net.Type,
			net.Managed,
			net.Config["ipv4.address"],
			net.Config["ipv6.address"],
			net.Description,
		)
	}
	return w.Flush()
}

func getNetwork(ctx context.Context, provider *incus.Provider, resource *resources.Resource, output string) error {
	network, err := provider.GetNetwork(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(network)
	}

	fmt.Printf("Name: %s\n", network.Name)
	fmt.Printf("Type: %s\n", network.Type)
	fmt.Printf("Managed: %v\n", network.Managed)
	fmt.Printf("Description: %s\n", network.Description)
	fmt.Printf("Used By: %v\n", network.UsedBy)

	return nil
}

func listStoragePools(ctx context.Context, provider *incus.Provider, namespace, output string) error {
	pools, err := provider.ListStoragePools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list storage pools: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pools)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDRIVER\tDESCRIPTION")
	for _, pool := range pools {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			pool.Name,
			pool.Driver,
			pool.Description,
		)
	}
	return w.Flush()
}

func getStoragePool(ctx context.Context, provider *incus.Provider, resource *resources.Resource, output string) error {
	pool, err := provider.GetStoragePool(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get storage pool: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pool)
	}

	fmt.Printf("Name: %s\n", pool.Name)
	fmt.Printf("Driver: %s\n", pool.Driver)
	fmt.Printf("Source: %s\n", pool.Config["source"])
	fmt.Printf("Description: %s\n", pool.Description)
	fmt.Printf("Used By: %v\n", pool.UsedBy)

	return nil
}

func listProfiles(ctx context.Context, provider *incus.Provider, namespace, output string) error {
	profiles, err := provider.ListProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(profiles)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	for _, profile := range profiles {
		fmt.Fprintf(w, "%s\t%s\n",
			profile.Name,
			profile.Description,
		)
	}
	return w.Flush()
}

func getProfile(ctx context.Context, provider *incus.Provider, resource *resources.Resource, output string) error {
	profile, err := provider.GetProfile(ctx, resource.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(profile)
	}

	fmt.Printf("Name: %s\n", profile.Name)
	fmt.Printf("Description: %s\n", profile.Description)
	fmt.Printf("Config: %v\n", profile.Config)
	fmt.Printf("Devices: %v\n", profile.Devices)
	fmt.Printf("Used By: %v\n", profile.UsedBy)

	return nil
}

func formatIPs(ips []string) string {
	if len(ips) == 0 {
		return "-"
	}
	if len(ips) == 1 {
		return ips[0]
	}
	return fmt.Sprintf("%s (+%d more)", ips[0], len(ips)-1)
}