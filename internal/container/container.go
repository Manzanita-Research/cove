package container

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func FindBinary() (string, error) {
	path, err := exec.LookPath("container")
	if err != nil {
		return "", fmt.Errorf("cove requires Apple's container CLI\nInstall it from https://github.com/apple/container/releases")
	}
	return path, nil
}

func SystemStatus(bin string) error {
	cmd := exec.Command(bin, "system", "status")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func SystemStart(bin string) error {
	cmd := exec.Command(bin, "system", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ImageExists(bin string, name string) (bool, error) {
	var out bytes.Buffer
	cmd := exec.Command(bin, "image", "list")
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false, err
	}
	return strings.Contains(out.String(), name), nil
}

func Build(bin string, tag string, contextDir string) error {
	cmd := exec.Command(bin, "build", "-t", tag, contextDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type RunOpts struct {
	Name       string
	Volumes    [][2]string // [host, container] pairs
	WorkDir    string
	Image      string
	Cmd        []string
}

func Run(bin string, opts RunOpts) error {
	args := []string{"run", "-it", "--rm"}

	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	for _, v := range opts.Volumes {
		args = append(args, "-v", v[0]+":"+v[1])
	}
	if opts.WorkDir != "" {
		args = append(args, "-w", opts.WorkDir)
	}
	args = append(args, opts.Image)
	args = append(args, opts.Cmd...)

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
