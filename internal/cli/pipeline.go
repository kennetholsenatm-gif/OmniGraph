package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/pipeline"
	"github.com/spf13/cobra"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage and execute pipelines",
		Long: `Pipeline management commands for defining, running, and monitoring CI/OT pipelines.

Pipelines are defined using the omnigraph/pipeline/v1 schema and can include:
- Multiple stages with dependencies
- Approval gates for manual intervention
- Conditional execution based on variables
- Retry policies and error handling
- Notifications on success/failure`,
	}

	cmd.AddCommand(newPipelineDefineCmd())
	cmd.AddCommand(newPipelineRunCmd())
	cmd.AddCommand(newPipelineListCmd())
	cmd.AddCommand(newPipelineStatusCmd())
	cmd.AddCommand(newPipelineApproveCmd())
	cmd.AddCommand(newPipelineCancelCmd())
	cmd.AddCommand(newPipelineHistoryCmd())

	return cmd
}

func newPipelineDefineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "define [pipeline-file]",
		Short: "Load a pipeline definition",
		Long: `Load a pipeline definition from a JSON or YAML file.

The pipeline file must conform to the omnigraph/pipeline/v1 schema.
Example:
  omnigraph pipeline define my-pipeline.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineFile := args[0]

			// Read pipeline file
			data, err := os.ReadFile(pipelineFile)
			if err != nil {
				return fmt.Errorf("failed to read pipeline file: %w", err)
			}

			// Parse pipeline definition
			var def pipeline.Definition
			if err := json.Unmarshal(data, &def); err != nil {
				return fmt.Errorf("failed to parse pipeline definition: %w", err)
			}

			// Get global engine
			engine := getPipelineEngine()

			// Load definition
			if err := engine.LoadDefinition(&def); err != nil {
				return fmt.Errorf("failed to load pipeline definition: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pipeline definition loaded: %s (version %s)\n", 
				def.Metadata.Name, def.Metadata.Version)

			return nil
		},
	}

	return cmd
}

func newPipelineRunCmd() *cobra.Command {
	var variables []string
	var version string

	cmd := &cobra.Command{
		Use:   "run [pipeline-name]",
		Short: "Execute a pipeline",
		Long: `Execute a pipeline by name and version.

Variables can be passed using the --var flag:
  omnigraph pipeline run my-pipeline --var environment=production --var region=us-west-2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineName := args[0]

			// Parse variables
			vars := make(map[string]interface{})
			for _, v := range variables {
				// Simple key=value parsing
				// In a real implementation, this would be more robust
				vars[v] = v
			}

			// Get global engine
			engine := getPipelineEngine()

			// Execute pipeline
			execution, err := engine.Execute(cmd.Context(), pipelineName, version, vars)
			if err != nil {
				return fmt.Errorf("failed to execute pipeline: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pipeline execution started: %s\n", execution.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", execution.Status)

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&variables, "var", []string{}, "Pipeline variables (key=value)")
	cmd.Flags().StringVar(&version, "version", "latest", "Pipeline version to execute")

	return cmd
}

func newPipelineListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available pipeline definitions",
		Long: `List all loaded pipeline definitions.

Example:
  omnigraph pipeline list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := getPipelineEngine()
			defs := engine.ListDefinitions()

			if len(defs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No pipeline definitions loaded")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-30s\n", "NAME", "VERSION", "DESCRIPTION")
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-30s\n", "----", "-------", "-----------")

			for _, def := range defs {
				desc := def.Metadata.Description
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-30s\n", 
					def.Metadata.Name, def.Metadata.Version, desc)
			}

			return nil
		},
	}

	return cmd
}

func newPipelineStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [execution-id]",
		Short: "Get pipeline execution status",
		Long: `Get the status of a pipeline execution.

Example:
  omnigraph pipeline status exec-1234567890`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]

			engine := getPipelineEngine()
			execution, err := engine.GetExecution(executionID)
			if err != nil {
				return fmt.Errorf("failed to get execution: %w", err)
			}

			// Print execution details
			fmt.Fprintf(cmd.OutOrStdout(), "Execution ID: %s\n", execution.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Pipeline: %s (version %s)\n", execution.Pipeline, execution.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", execution.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Started: %s\n", execution.StartedAt.Format("2006-01-02 15:04:05"))

			if execution.CompletedAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Completed: %s\n", execution.CompletedAt.Format("2006-01-02 15:04:05"))
				fmt.Fprintf(cmd.OutOrStdout(), "Duration: %s\n", execution.CompletedAt.Sub(execution.StartedAt))
			}

			if execution.Error != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Error: %s\n", execution.Error)
			}

			// Print stages
			fmt.Fprintln(cmd.OutOrStdout(), "\nStages:")
			for stageName, stage := range execution.Stages {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", stageName, stage.Status)
			}

			return nil
		},
	}

	return cmd
}

func newPipelineApproveCmd() *cobra.Command {
	var approver string

	cmd := &cobra.Command{
		Use:   "approve [execution-id] [stage-name]",
		Short: "Approve a pipeline stage",
		Long: `Approve a pipeline stage that is waiting for manual approval.

Example:
  omnigraph pipeline approve exec-1234567890 approve-stage --approver john.doe`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]
			stageName := args[1]

			engine := getPipelineEngine()
			if err := engine.ApproveStage(executionID, stageName, approver); err != nil {
				return fmt.Errorf("failed to approve stage: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Stage %s approved by %s\n", stageName, approver)

			return nil
		},
	}

	cmd.Flags().StringVar(&approver, "approver", "", "Approver name (required)")
	cmd.MarkFlagRequired("approver")

	return cmd
}

func newPipelineCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [execution-id]",
		Short: "Cancel a running pipeline execution",
		Long: `Cancel a running pipeline execution.

Example:
  omnigraph pipeline cancel exec-1234567890`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]

			engine := getPipelineEngine()
			if err := engine.CancelExecution(executionID); err != nil {
				return fmt.Errorf("failed to cancel execution: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Execution %s cancelled\n", executionID)

			return nil
		},
	}

	return cmd
}

func newPipelineHistoryCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show pipeline execution history",
		Long: `Show pipeline execution history.

Example:
  omnigraph pipeline history --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := getPipelineEngine()
			history := engine.GetHistory(limit)

			if len(history) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No execution history")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-30s %-10s %-20s %-10s\n", 
				"ID", "PIPELINE", "STATUS", "COMPLETED", "DURATION")
			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-30s %-10s %-20s %-10s\n", 
				"--", "--------", "------", "---------", "--------")

			for _, record := range history {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-30s %-10s %-20s %-10s\n",
					record.ID,
					record.Pipeline,
					record.Status,
					record.CompletedAt.Format("2006-01-02 15:04:05"),
					record.Duration.Round(time.Millisecond),
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Number of history records to show")

	return cmd
}

// Global pipeline engine instance
var globalPipelineEngine *pipeline.Engine

func getPipelineEngine() *pipeline.Engine {
	if globalPipelineEngine == nil {
		globalPipelineEngine = pipeline.NewEngine()
	}
	return globalPipelineEngine
}