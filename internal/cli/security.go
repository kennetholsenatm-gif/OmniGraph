package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
	"github.com/spf13/cobra"
)

func newSecurityCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "security",
		Short: "Passive security posture scans (authorized validation only)",
		Long: `Run ATT&CK-aligned, read-only checks against Linux targets.

Use only on systems you own or are explicitly authorized to test. Misuse may be unlawful.

Examples:
  omnigraph security scan --local --output scan.json
  omnigraph security scan --inventory hosts.ini --ssh-user ec2-user --ssh-key ~/.ssh/id_rsa --ssh-known-hosts ~/.ssh/known_hosts --output-dir ./scans`,
	}
	root.AddCommand(newSecurityScanCmd(), newSecurityListModulesCmd())
	return root
}

func newSecurityListModulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-modules",
		Short: "List built-in posture modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, m := range security.All {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", m.ID(), m.TechniqueID(), m.Tactic(), m.TechniqueName())
			}
			return nil
		},
	}
}

func newSecurityScanCmd() *cobra.Command {
	var (
		local, insecureSkipHostKey bool
		inventoryPath             string
		sshUser, sshKey, sshPort  string
		sshKnownHosts             string
		limit                     int
		outputPath, outputDir     string
		profile                   string
		tactic, technique, module string
		timeout                   time.Duration
	)
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run posture modules and write omnigraph/security/v1 JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			f := security.Filter{Tactic: tactic, Technique: technique, ModuleID: module}
			if (local && inventoryPath != "") || (!local && inventoryPath == "") {
				return fmt.Errorf("specify exactly one of --local or --inventory")
			}
			if inventoryPath != "" && strings.TrimSpace(sshUser) == "" {
				return fmt.Errorf("--ssh-user is required with --inventory")
			}
			if inventoryPath != "" && sshKey == "" {
				return fmt.Errorf("--ssh-key is required with --inventory (or set a default key path)")
			}
			if outputPath != "" && outputDir != "" {
				return fmt.Errorf("use either --output or --output-dir, not both")
			}
			if inventoryPath != "" && outputPath != "" {
				return fmt.Errorf("--inventory requires --output-dir (multiple hosts)")
			}
			if local {
				if outputPath == "" {
					return fmt.Errorf("--output is required for --local")
				}
				h := security.LocalHost{}
				doc := security.Run(ctx, h, "local", profile, "", f, timeout)
				b, err := security.EncodeIndent(doc)
				if err != nil {
					return err
				}
				return os.WriteFile(outputPath, append(b, '\n'), 0o644)
			}
			raw, err := os.ReadFile(inventoryPath)
			if err != nil {
				return err
			}
			hosts := security.ParseAnsibleInventoryINI(string(raw))
			if len(hosts) == 0 {
				return fmt.Errorf("no hosts parsed from %s", inventoryPath)
			}
			if outputDir == "" {
				return fmt.Errorf("--output-dir is required for --inventory")
			}
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return err
			}
			n := 0
			for _, inv := range hosts {
				if limit > 0 && n >= limit {
					break
				}
				cfg := security.SSHDialConfig{
					Host:            inv.Host,
					Port:            firstNonEmpty(inv.Port, sshPort, "22"),
					User:            firstNonEmpty(inv.User, sshUser),
					KeyPath:         sshKey,
					InsecureHostKey: insecureSkipHostKey,
					KnownHostsPath:  sshKnownHosts,
				}
				sh, err := security.DialSSH(cfg)
				if err != nil {
					return fmt.Errorf("ssh %s: %w", inv.Host, err)
				}
				doc := security.Run(ctx, sh, "ssh", profile, inv.Host, f, timeout)
				_ = sh.Close()
				b, err := security.EncodeIndent(doc)
				if err != nil {
					return err
				}
				name := security.SanitizeFilename(inv.Name + "_" + inv.Host)
				out := filepath.Join(outputDir, "scan_"+name+".json")
				if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", out)
				n++
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "scan the machine running omnigraph (Linux targets only for module relevance)")
	cmd.Flags().StringVar(&inventoryPath, "inventory", "", "Ansible INI inventory path for SSH targets")
	cmd.Flags().StringVar(&sshUser, "ssh-user", "", "SSH username (or per-host ansible_user in inventory)")
	cmd.Flags().StringVar(&sshKey, "ssh-key", "", "path to SSH private key")
	cmd.Flags().StringVar(&sshPort, "ssh-port", "", "default SSH port when not set per host (default 22)")
	cmd.Flags().StringVar(&sshKnownHosts, "ssh-known-hosts", "", "known_hosts file for host key verification")
	cmd.Flags().BoolVar(&insecureSkipHostKey, "insecure-ignore-host-key", false, "disable SSH host key verification (unsafe)")
	cmd.Flags().IntVar(&limit, "limit", 0, "max hosts from inventory (0 = all)")
	cmd.Flags().StringVar(&outputPath, "output", "", "write single scan JSON here (--local)")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "write one JSON file per inventory host")
	cmd.Flags().StringVar(&profile, "profile", "default", "profile label stored in scan metadata")
	cmd.Flags().StringVar(&tactic, "tactic", "", "run only modules for this tactic (e.g. defense_evasion)")
	cmd.Flags().StringVar(&technique, "technique", "", "run only modules for this technique id (e.g. T1082)")
	cmd.Flags().StringVar(&module, "module", "", "run only this module id (e.g. selinux_mode)")
	cmd.Flags().DurationVar(&timeout, "module-timeout", 90*time.Second, "timeout per module")
	return cmd
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
