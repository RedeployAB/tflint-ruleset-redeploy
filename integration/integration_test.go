package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
		{
			Name:    "monorepo-module",
			Command: exec.Command("tflint", "--format", "json", "--force"),
			Dir:     "monorepo-module",
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

func TestIntegrationAutofix(t *testing.T) {
	cases := []struct {
		Name         string
		Dir          string
		InputFile    string
		ExpectedFile string
	}{
		{
			Name:         "autofix preserves comments before blocks",
			Dir:          "autofix-meta-order-comments",
			InputFile:    "main.tf",
			ExpectedFile: "expected.tf",
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			srcDir := filepath.Join(dir, tc.Dir)
			tmpDir := copyFixtureDir(t, srcDir, tc.ExpectedFile)

			t.Cleanup(func() {
				if chdirErr := os.Chdir(dir); chdirErr != nil {
					t.Fatal(chdirErr)
				}
			})
			if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
				t.Fatal(chdirErr)
			}

			stdout := runTflint(t, "--fix", "--force", "--format", "json")
			assertJSONMatch(t, stdout, filepath.Join(srcDir, "result.json"))
			assertFileMatch(t, filepath.Join(tmpDir, tc.InputFile), filepath.Join(srcDir, tc.ExpectedFile))
		})
	}
}

// copyFixtureDir copies all files (except expectedFile and result.json) from srcDir to a temp dir.
func copyFixtureDir(t *testing.T, srcDir, expectedFile string) string {
	t.Helper()
	tmpDir := t.TempDir()
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("reading dir %s: %v", srcDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == expectedFile || entry.Name() == "result.json" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(srcDir, entry.Name()))
		if readErr != nil {
			t.Fatalf("reading %s: %v", entry.Name(), readErr)
		}
		if writeErr := os.WriteFile(filepath.Join(tmpDir, entry.Name()), data, 0o644); writeErr != nil {
			t.Fatalf("writing %s: %v", entry.Name(), writeErr)
		}
	}
	return tmpDir
}

func runTflint(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("tflint", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s, stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	return stdout.Bytes()
}

func assertJSONMatch(t *testing.T, stdout []byte, resultPath string) {
	t.Helper()
	resultJSON, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("reading %s: %v", resultPath, err)
	}
	var expected, got interface{}
	if err := json.Unmarshal(resultJSON, &expected); err != nil {
		t.Fatalf("unmarshaling %s: %v", resultPath, err)
	}
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshaling stdout: %v", err)
	}
	// Strip fixable/fixed fields that vary between tflint versions.
	stripIssueFields(got)
	stripIssueFields(expected)
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Fatalf("JSON output mismatch (-got +want):\n%s", diff)
	}
}

// stripIssueFields removes version-dependent fields (fixable, fixed) from tflint JSON output.
func stripIssueFields(v interface{}) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return
	}
	issues, ok := m["issues"].([]interface{})
	if !ok {
		return
	}
	for _, issue := range issues {
		if im, ok := issue.(map[string]interface{}); ok {
			delete(im, "fixable")
			delete(im, "fixed")
		}
	}
}

func assertFileMatch(t *testing.T, gotPath, expectedPath string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("reading %s: %v", gotPath, err)
	}
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("reading %s: %v", expectedPath, err)
	}
	// Normalize line endings for Windows compatibility.
	gotStr := strings.ReplaceAll(string(got), "\r\n", "\n")
	expectedStr := strings.ReplaceAll(string(expected), "\r\n", "\n")
	if diff := cmp.Diff(gotStr, expectedStr); diff != "" {
		t.Fatalf("fixed file does not match expected (-got +want):\n%s", diff)
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
