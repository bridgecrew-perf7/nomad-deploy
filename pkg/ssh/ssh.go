package ssh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/config"
)

// Ssh simply executes any shell command on remote host
func Ssh(host config.Host, cfg *config.Config, command string) (string, error) {
	cmd := exec.Command("ssh",
		"-p", strconv.Itoa(int(host.SshPort)),
		"-i", cfg.SSHKey,
		fmt.Sprintf("%s@%s", host.User, host.Address),
		command)
	cmd.Stderr = os.Stderr

	output := bytes.Buffer{}
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}

// Scp simply copies local file to remote host
func Scp(host config.Host, cfg *config.Config, localPath, remotePath string) error {
	cmd := exec.Command("scp",
		"-P", strconv.Itoa(int(host.SshPort)),
		"-i", cfg.SSHKey,
		localPath,
		fmt.Sprintf("%s@%s:%s", host.User, host.Address, remotePath))
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
