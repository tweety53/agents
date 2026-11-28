package core

import (
	"bytes"
	"os/exec"
)

// RunCombined runs name args from dir (or env's directory), returning its exit
// code and combined output.
func RunCombined(env Env, dir, name string, args ...string) (int, string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = env.Dir
	if dir != "" {
		cmd.Dir = dir
	}
	return Combined(cmd)
}

// Combined runs cmd, returning its exit code and combined output.
func Combined(cmd *exec.Cmd) (int, string) {
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return ExitCode(cmd.Run()), out.String()
}
