//go:build windows

package scenariolog

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {
	// No-op on Windows for now.
	_ = cmd
}

func terminateProcessGroupBestEffort(pgid int) {
	// No-op on Windows for now.
	_ = pgid
}


