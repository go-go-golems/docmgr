//go:build !windows

package scenariolog

import (
	"os/exec"
	"syscall"
	"time"
)

func setProcessGroup(cmd *exec.Cmd) {
	// Create a new process group so we can terminate the whole subtree.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcessGroupBestEffort(pgid int) {
	// Negative pid targets the process group on unix.
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	time.Sleep(250 * time.Millisecond)
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
}


