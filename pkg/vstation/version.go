package vstation

import (
	"bytes"
	"os/exec"
)

// vstationStd ...
func vstationStd(arg ...string) ([]byte, error) {
	var (
		cmd    *exec.Cmd
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd = exec.Command(vstationName, arg...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), nil
}

// Version - get vit-servicing-station-server version.
//
//  vit-servicing-station-server --version | STDOUT
func Version() ([]byte, error) {
	return vstationStd("--version")
}
