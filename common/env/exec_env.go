package env

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SubprocessEnv returns os.Environ() enriched with standard compiler/binary directories.
func SubprocessEnv() []string {
	env := os.Environ()
	if runtime.GOOS == "windows" {
		userProfile := os.Getenv("USERPROFILE")
		extraPaths := []string{
			`C:\Program Files\Go\bin`,
			filepath.Join(userProfile, "go", "bin"),
			filepath.Join(userProfile, ".agentfence", "bin"),
		}
		for i, e := range env {
			if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
				currPath := e[5:]
				newPath := strings.Join(extraPaths, ";") + ";" + currPath
				env[i] = "PATH=" + newPath
				return env
			}
		}
		env = append(env, "PATH="+strings.Join(extraPaths, ";"))
	}
	return env
}
