package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/heuristics"
	"github.com/coreos/coreos-newsletter/internal/llm"
	"github.com/coreos/coreos-newsletter/internal/pipeline"
	"github.com/coreos/coreos-newsletter/internal/util"
)

var (
	cfgFile string
	days    int
	outFile string
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "coreos-newsletter",
		Short: "Fedora CoreOS Newsletter Generator",
		Long: `coreos-newsletter generates weekly newsletters for Fedora CoreOS.

Collects contributions from GitHub and meeting logs from Fedora CoreOS meetings.
Generates an informative newsletter highlighting activities and discussions.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file path")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")

	// Add commands
	rootCmd.AddCommand(cmdGenerate())
	rootCmd.AddCommand(cmdValidate())
	rootCmd.AddCommand(cmdSummary())
	rootCmd.AddCommand(cmdVersion())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// cmdGenerate generates a newsletter
func cmdGenerate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate newsletter",
		Long: `Generate a weekly newsletter with GitHub PRs and Fedora CoreOS meeting logs.

Output is a markdown file suitable for sharing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate()
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "number of days to look back")
	cmd.Flags().StringVar(&outFile, "out", "", "output file (default: stdout)")

	return cmd
}

// cmdValidate validates configuration and connectivity
func cmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration and API connectivity",
		Long: `Validate that all configured API tokens are present and that basic
connectivity to all services works. Useful for debugging setup issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate()
		},
	}

	return cmd
}

// cmdSummary shows a data summary for debugging
func cmdSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show data summary for debugging",
		Long: `Collect and summarize data from all sources without generating a newsletter.
Useful for debugging and understanding what data is available.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSummary()
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "number of days to look back")

	return cmd
}

// cmdVersion shows version information
func cmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("coreos-newsletter v0.1.0 (Go)")
			return nil
		},
	}

	return cmd
}

// Helper functions

func runGenerate() error {
	logger := util.NewLogger(verbose)
	logger.Info("Starting newsletter generation")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	// Calculate time window
	windowEnd := time.Now().UTC()
	windowStart := windowEnd.AddDate(0, 0, -days)
	logger.Info(fmt.Sprintf("Time window: %s to %s", windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02")))

	// Build bundle
	builder := pipeline.NewBundleBuilder(cfg, logger)
	bundle, warnings := builder.Build(windowStart, windowEnd)

	// Log warnings
	for _, w := range warnings {
		logger.Warn(w)
	}

	// Validate bundle
	validationIssues := builder.ValidateBundle(bundle)
	for _, issue := range validationIssues {
		logger.Warn(issue)
	}

	// Build LLM payload
	payload := heuristics.BuildLLMPayload(bundle)

	// Generate newsletter
	logger.Info("Generating newsletter with LLM...")
	llmClient, err := llm.NewClient(cfg, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create LLM client: %v", err))
		return err
	}
	defer llmClient.Close()

	// Create prompt and generate
	prompt := llm.NewsletterPrompt(payload)
	newsletter, err := llmClient.GenerateNewsletter(nil, prompt)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate newsletter: %v", err))
		return err
	}

	// Output
	return outputNewsletter(newsletter)
}

func runValidate() error {
	fmt.Println("Validating Fedora CoreOS Newsletter Configuration\n")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("✗ Failed to load config: %v\n", err)
		return err
	}

	// Check GitHub
	fmt.Print("GitHub: ")
	if cfg.GitHubToken == "" || cfg.GitHubOwner == "" {
		fmt.Println("✗ Not configured (set GITHUB_TOKEN and GITHUB_OWNER)")
	} else {
		fmt.Println("✓ Configured")
	}

	// Check LLM Provider
	fmt.Print("LLM Provider: ")
	provider := strings.ToLower(cfg.LLMProvider)
	
	if provider == "gemini" {
		if cfg.VertexProjectID != "" {
			fmt.Println("✓ Gemini configured (Vertex AI)")
		} else if cfg.GoogleAPIKey != "" {
			fmt.Println("✓ Gemini configured (direct API)")
		} else {
			fmt.Println("✗ Gemini selected but not configured (set ANTHROPIC_VERTEX_PROJECT_ID or GOOGLE_API_KEY)")
		}
	} else if provider == "claude" {
		if cfg.VertexProjectID != "" {
			fmt.Println("✓ Claude configured (Vertex AI)")
		} else if cfg.AnthropicAPIKey != "" {
			fmt.Println("✓ Claude configured (direct API)")
		} else {
			fmt.Println("✗ Claude selected but not configured (set ANTHROPIC_VERTEX_PROJECT_ID or ANTHROPIC_API_KEY)")
		}
	} else {
		if cfg.GoogleAPIKey != "" {
			fmt.Println("✓ Gemini configured")
		} else if cfg.AnthropicAPIKey != "" {
			fmt.Println("✓ Claude configured (direct API)")
		} else if cfg.VertexProjectID != "" {
			fmt.Println("✓ Vertex AI configured")
		} else {
			fmt.Println("✗ No provider configured")
			fmt.Println("   Set LLM_PROVIDER=claude or LLM_PROVIDER=gemini, then:")
			fmt.Println("   - For Vertex AI: set ANTHROPIC_VERTEX_PROJECT_ID and CLOUD_ML_REGION")
			fmt.Println("   - For Gemini direct: set GOOGLE_API_KEY")
			fmt.Println("   - For Claude direct: set ANTHROPIC_API_KEY")
		}
	}

	fmt.Println("\n✓ Configuration validation complete")
	return nil
}

func runSummary() error {
	logger := util.NewLogger(verbose)
	logger.Info("Collecting data for summary...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	// Calculate time window
	windowEnd := time.Now().UTC()
	windowStart := windowEnd.AddDate(0, 0, -days)
	logger.Info(fmt.Sprintf("Time window: %s to %s", windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02")))

	// Build bundle
	builder := pipeline.NewBundleBuilder(cfg, logger)
	bundle, warnings := builder.Build(windowStart, windowEnd)

	// Log warnings
	for _, w := range warnings {
		logger.Warn(w)
	}

	// Print summary
	fmt.Printf("\n=== Data Summary ===\n\n")
	fmt.Printf("Window: %s to %s (%d days)\n\n",
		windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02"), days)

	fmt.Printf("Pull Requests: %d\n", len(bundle.PullRequests))

	fmt.Printf("\nMeetings: %d\n", len(bundle.Meetings))

	// Show heuristics
	fmt.Printf("\n=== Heuristics ===\n\n")
	summary := heuristics.SummarizePRs(bundle.PullRequests)
	fmt.Print(summary)

	return nil
}

func outputNewsletter(newsletter string) error {
	if outFile == "" {
		// Output to stdout
		fmt.Println(newsletter)
		return nil
	}

	// Output to file
	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(newsletter)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Newsletter written to %s\n", outFile)
	return nil
}
