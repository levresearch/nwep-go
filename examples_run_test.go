package nwep_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestExamplesRun runs each standalone example program and requires a clean exit
// with output, so an example that rots is a failing test NWG1100. the
// examples self-verify with asserts, the zero exit is the contract.
func TestExamplesRun(t *testing.T) {
	for _, name := range []string{"managed", "managed_dht", "managed_stream"} {
		t.Run(name, func(t *testing.T) {
			out, err := exec.Command("go", "run", "./examples/"+name+"/").CombinedOutput() //nolint:gosec // name is from a hard-coded literal slice, not user input
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", name, err, out)
			}
			if !strings.Contains(string(out), "shutdown       clean") {
				t.Fatalf("%s did not finish cleanly:\n%s", name, out)
			}
		})
	}
}
