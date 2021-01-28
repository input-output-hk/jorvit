// Package vcli provides vit-servicing-station-cli binary helpers.
package vcli

import (
	"bytes"
	"os/exec"
)

var (
	vcliName = "vit-servicing-station-cli"
)

// vcli executes "stdin | 'vcliName' args | stdout"
func vcli(stdin []byte, arg ...string) ([]byte, error) {
	var (
		cmd    *exec.Cmd
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd = exec.Command(vcliName, arg...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != nil /* && len(stdin) > 0 */ {
		cmd.Stdin = bytes.NewBuffer(stdin)
	}

	if err := cmd.Run(); err != nil {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), nil
}

// BinName set the executable name/path if not the default one.
func BinName(name string) {
	vcliName = name
}
