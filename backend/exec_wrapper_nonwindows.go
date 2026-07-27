//go:build !windows

package backend

import "os/exec"

// applyHiddenWindowAttr is a no-op on non-Windows platforms.
func applyHiddenWindowAttr(cmd *exec.Cmd) {
	// Non-Windows processes do not need hidden-window attributes.
}
