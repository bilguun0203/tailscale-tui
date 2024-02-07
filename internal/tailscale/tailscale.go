package tailscale

import (
	"encoding/json"
	"os/exec"
	"runtime"
)

type Tailscale struct {
	cliPath string
}

func (ts Tailscale) runCommand(arg ...string) ([]byte, error) {
	cmd := exec.Command(ts.cliPath, arg...)
	output, err := cmd.Output()
	return output, err
}

func (ts Tailscale) Status() (Status, error) {
	output, err := ts.runCommand("status", "--json")
	status := Status{}
	if err != nil {
		return status, err
	}
	err = json.Unmarshal(output, &status)
	if err != nil {
		return status, err
	}
	return status, nil
}

func New() (Tailscale, error) {
	cliPath := "tailscale"
	if runtime.GOOS == "darwin" {
		cliPath = "/Applications/Tailscale.app/Contents/MacOS/Tailscale"
	}
	_, err := exec.LookPath(cliPath)
	return Tailscale{cliPath: cliPath}, err
}
