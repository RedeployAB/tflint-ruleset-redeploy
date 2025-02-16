package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command *exec.Cmd
		Dir     string
	}{
		{
			Name:    "basic-module",
			Command: exec.Command("tflint", "--format", "json", "--force"),
			Dir:     "basic-module",
		},
		{
			Name:    "valid-module",
			Command: exec.Command("tflint", "--format", "json", "--force"),
			Dir:     "valid-module",
		},
		{
			Name:    "invalid-module",
			Command: exec.Command("tflint", "--format", "json", "--force"),
			Dir:     "invalid-module",
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := filepath.Join(dir, tc.Dir)

			t.Cleanup(func() {
				if chdirErr := os.Chdir(dir); chdirErr != nil {
					t.Fatal(chdirErr)
				}
			})

			if chdirErr := os.Chdir(testDir); chdirErr != nil {
				t.Fatal(chdirErr)
			}

			var stdout, stderr bytes.Buffer
			tc.Command.Stdout = &stdout
			tc.Command.Stderr = &stderr
			if cmdErr := tc.Command.Run(); cmdErr != nil {
				t.Fatalf("%s, stdout=%s stderr=%s", cmdErr, stdout.String(), stderr.String())
			}

			var b []byte
			if runtime.GOOS == "windows" && IsWindowsResultExist() {
				b, err = os.ReadFile(filepath.Join(testDir, "result_windows.json"))
			} else {
				b, err = os.ReadFile(filepath.Join(testDir, "result.json"))
			}
			if err != nil {
				t.Fatal(err)
			}

			var expected interface{}
			if err := json.Unmarshal(b, &expected); err != nil {
				t.Fatal(err)
			}

			var got interface{}
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
