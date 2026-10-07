//go:build !linux

package bkjw

import "os/exec"

func configureCollectorProcess(cmd *exec.Cmd) {}
func terminateCollectorProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
