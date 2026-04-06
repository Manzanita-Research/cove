package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manzanita-research/cove/internal/banner"
	"github.com/manzanita-research/cove/internal/container"
	"github.com/spf13/cobra"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)

var (
	Version = "dev"

	rebuild bool

	embedFS embed.FS
)

const imageName = "cove:latest"

var rootCmd = &cobra.Command{
	Use:   "cove",
	Short: "Sandboxed Claude Code sessions using Apple Containers",
	Long:  "Drop into an ephemeral Linux microVM where Claude Code runs with --dangerously-skip-permissions, scoped to your current directory.",
	RunE:  run,
}

func init() {
	rootCmd.Flags().BoolVar(&rebuild, "rebuild", false, "Force rebuild the container image")
	rootCmd.Version = Version
}

func Execute(fs embed.FS) {
	embedFS = fs
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	bin, err := container.FindBinary()
	if err != nil {
		return err
	}

	// ensure container system is running
	if err := container.SystemStatus(bin); err != nil {
		banner.Dim("starting container system...")
		if err := container.SystemStart(bin); err != nil {
			return fmt.Errorf("failed to start container system: %w", err)
		}
	}

	// ensure kernel is configured
	if !container.KernelConfigured(bin) {
		banner.Dim("installing default kernel...")
		if err := container.KernelSet(bin); err != nil {
			return fmt.Errorf("failed to install kernel: %w", err)
		}
	}

	// build image if needed
	exists, err := container.ImageExists(bin, imageName)
	if err != nil {
		return fmt.Errorf("failed to check images: %w", err)
	}

	if !exists || rebuild {
		banner.Warm("building cove image...")

		tmpDir, err := writeDockerfile()
		if err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		if err := container.Build(bin, imageName, tmpDir); err != nil {
			return fmt.Errorf("image build failed (try cove --rebuild to retry): %w", err)
		}
		banner.Warm("image built")
	}

	// resolve paths
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	project := sanitizeName(filepath.Base(cwd))

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create ~/.claude: %w", err)
	}

	// copy .claude.json into .claude/ so we only need one directory mount
	// (apple containers doesn't support single-file volume mounts)
	claudeJSON := filepath.Join(home, ".claude.json")
	if data, err := os.ReadFile(claudeJSON); err == nil {
		os.WriteFile(filepath.Join(claudeDir, ".claude.json"), data, 0644)
	}

	// print banner
	fmt.Println()
	banner.Warm("entering cove for: " + project)
	banner.Dim("  mounted: " + cwd + " -> /workspace")
	banner.Dim("  auth:    ~/.claude (shared with host)")
	banner.Dim("  exit claude to destroy the sandbox")
	fmt.Println()

	// run it
	return container.Run(bin, container.RunOpts{
		Name:    fmt.Sprintf("cove-%s-%d", project, os.Getpid()),
		Volumes: [][2]string{
			{cwd, "/workspace"},
			{claudeDir, "/home/cove/.claude"},
		},
		WorkDir: "/workspace",
		Image:   imageName,
		Cmd:     []string{"claude", "--dangerously-skip-permissions"},
	})
}

func writeDockerfile() (string, error) {
	tmpDir, err := os.MkdirTemp("", "cove-build-*")
	if err != nil {
		return "", err
	}

	data, err := embedFS.ReadFile("embed/Dockerfile")
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), data, 0644); err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	entrypoint, err := embedFS.ReadFile("embed/entrypoint.sh")
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "entrypoint.sh"), entrypoint, 0755); err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return tmpDir, nil
}

func sanitizeName(name string) string {
	name = nonAlphanumeric.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "project"
	}
	return name
}
