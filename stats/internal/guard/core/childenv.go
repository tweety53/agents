package core

import (
	"os"
)

// ChildEnv is what every child runs under: the environment the caller
// handed over, the process's own in production.
func ChildEnv(env Env) []string {
	if env.Environ != nil {
		return env.Environ()
	}
	return os.Environ()
}

// IsRegular is `[ -f <p> ]`.
func IsRegular(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}
