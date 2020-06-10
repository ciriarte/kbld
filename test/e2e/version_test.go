package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

	out, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})

	if !strings.Contains(out, "kbld version") {
		t.Fatalf("Expected to find client version")
	}
}
