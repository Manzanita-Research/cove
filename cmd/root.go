package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manzanita-research/claudebox/internal/banner"
	"github.com/manzanita-research/claudebox/internal/container"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"

	rebuild bool

	embedFS embed.FS
)

const imageName = "claudebox:latest"

var rootCmd = &cobra.Command{
	Use:   "claudebox",
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

	// build image if needed
	exists, err := container.ImageExists(bin, "claudebox")
	if err != nil {
		return fmt.Errorf("failed to check images: %w", err)
	}

	if !exists || rebuild {
		banner.Warm("building claudebox image...")

		tmpDir, err := writeDockerfile()
		if err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		if err := container.Build(bin, imageName, tmpDir); err != nil {
			return fmt.Errorf("image build failed: %w", err)
		}
		banner.Warm("image built")
	}

	// resolve paths
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	project := filepath.Base(cwd)

	claudeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	claudeDir = filepath.Join(claudeDir, ".claude")

	// print banner
	fmt.Println()
	banner.Warm("entering claudebox for: " + project)
	banner.Dim("  mounted: " + cwd + " -> /workspace")
	banner.Dim("  auth:    ~/.claude (shared with host)")
	banner.Dim("  exit claude to destroy the sandbox")
	fmt.Println()

	// run it
	return container.Run(bin, container.RunOpts{
		Name:    fmt.Sprintf("claudebox-%s-%d", project, os.Getpid()),
		Volumes: [][2]string{{cwd, "/workspace"}, {claudeDir, "/root/.claude"}},
		WorkDir: "/workspace",
		Image:   imageName,
		Cmd:     []string{"claude", "--dangerously-skip-permissions"},
	})
}

func writeDockerfile() (string, error) {
	tmpDir, err := os.MkdirTemp("", "claudebox-build-*")
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

	return tmpDir, nil
}
