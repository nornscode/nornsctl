package dev

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// docker runs a docker command and returns combined output.
// Child processes are placed in their own process group so terminal
// SIGINT doesn't kill them (important during Ctrl-C shutdown).
func docker(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// dockerRun starts a container. Returns container ID.
func dockerRun(args []string) (string, error) {
	full := append([]string{"run", "-d"}, args...)
	return docker(full...)
}

// CheckDockerAvailable verifies the docker CLI is installed and the daemon is running.
func CheckDockerAvailable() error {
	path, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker not found in PATH. Install Docker: https://docs.docker.com/get-docker/")
	}
	_ = path

	out, err := docker("info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return fmt.Errorf("docker daemon is not running. Start Docker and try again.\n\n%s", out)
	}
	return nil
}

// IsRunning checks if a container with the given name is running.
func IsRunning(name string) bool {
	out, err := docker("inspect", "-f", "{{.State.Running}}", name)
	if err != nil {
		return false
	}
	return out == "true"
}

// ContainerExists checks if a container with the given name exists (running or stopped).
func ContainerExists(name string) bool {
	_, err := docker("inspect", name)
	return err == nil
}

// RemoveContainer force-removes a container by name.
func RemoveContainer(name string) {
	docker("rm", "-f", name)
}

// StopContainer stops a container by name.
func StopContainer(name string) {
	docker("stop", name)
}

// PullImage pulls a Docker image, streaming progress to stdout.
func PullImage(img string) error {
	fmt.Printf("Pulling %s...\n", img)
	cmd := exec.Command("docker", "pull", img)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// NetworkExists checks if a Docker network exists.
func NetworkExists(name string) bool {
	out, err := docker("network", "ls", "--filter", "name="+name, "--format", "{{.Name}}")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

// CreateNetwork creates a Docker network.
func CreateNetwork(name string) error {
	_, err := docker("network", "create", name)
	return err
}

// RemoveNetwork removes a Docker network.
func RemoveNetwork(name string) {
	docker("network", "rm", name)
}

// VolumeExists checks if a Docker volume exists.
func VolumeExists(name string) bool {
	out, err := docker("volume", "ls", "--filter", "name="+name, "--format", "{{.Name}}")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

// CreateVolume creates a Docker volume.
func CreateVolume(name string) error {
	_, err := docker("volume", "create", name)
	return err
}

// RemoveVolume removes a Docker volume.
func RemoveVolume(name string) error {
	_, err := docker("volume", "rm", name)
	return err
}

// StreamLogs tails logs from a container to stdout. Blocks until interrupted.
func StreamLogs(name string) error {
	cmd := exec.Command("docker", "logs", "-f", "--tail", "50", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
