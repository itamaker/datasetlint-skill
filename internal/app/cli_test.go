package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests swap os.Stdout/os.Stderr, so none of them call t.Parallel.

func runCapturing(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()

	code = Run(args)

	os.Stdout, os.Stderr = oldOut, oldErr
	_ = outW.Close()
	_ = errW.Close()
	read := func(f *os.File) string {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := f.Read(buf)
			sb.Write(buf[:n])
			if err != nil {
				break
			}
		}
		_ = f.Close()
		return sb.String()
	}
	return code, read(outR), read(errR)
}

func TestHelpAndVersionExitZero(t *testing.T) {
	cases := [][]string{
		{"-h"}, {"-help"}, {"--help"}, {"help"}, {"help", "-h"}, {"help", "--help"},
		{"help", "scan"}, {"help", "tui"}, {"help", "interactive"},
		{"scan", "-h"}, {"scan", "-help"}, {"scan", "--help"}, {"scan", "help"},
		{"tui", "-h"}, {"tui", "--help"}, {"tui", "help"},
		{"interactive", "--help"}, {"interactive", "help"},
		{"version"}, {"-version"}, {"--version"},
		{"scan", "version"}, {"scan", "-version"}, {"scan", "--version"},
		{"tui", "version"}, {"tui", "--version"},
		{"interactive", "version"}, {"interactive", "--version"},
	}
	for _, args := range cases {
		code, stdout, stderr := runCapturing(t, args...)
		if code != 0 {
			t.Errorf("%v: exit code = %d, want 0 (stderr %q)", args, code, stderr)
		}
		if strings.TrimSpace(stdout) == "" {
			t.Errorf("%v: expected output on stdout", args)
		}
		if stderr != "" {
			t.Errorf("%v: unexpected stderr %q", args, stderr)
		}
	}
}

func TestHelpContent(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"--help"}, []string{"datasetlint scan", "datasetlint version", "must not", "Exit codes"}},
		{[]string{"scan", "-h"}, []string{"Usage: datasetlint scan", "-train", "-eval", "-semantic-threshold", "-max-samples", "-strict", "-json"}},
		{[]string{"tui", "--help"}, []string{"Usage: datasetlint", "datasetlint scan"}},
		{[]string{"--version"}, []string{"datasetlint "}},
	}
	for _, tc := range cases {
		_, stdout, _ := runCapturing(t, tc.args...)
		for _, want := range tc.want {
			if !strings.Contains(stdout, want) {
				t.Errorf("%v: output is missing %q:\n%s", tc.args, want, stdout)
			}
		}
	}
}

func TestHelpSpellingsAgree(t *testing.T) {
	_, top, _ := runCapturing(t, "help")
	for _, args := range [][]string{{"-h"}, {"-help"}, {"--help"}, {"help", "-h"}, {"help", "--help"}, {"help", "help"}} {
		if _, out, _ := runCapturing(t, args...); out != top {
			t.Errorf("%v: output differs from top-level help", args)
		}
	}

	_, scanHelp, _ := runCapturing(t, "scan", "-h")
	for _, args := range [][]string{{"scan", "--help"}, {"scan", "help"}, {"help", "scan"}} {
		if _, out, _ := runCapturing(t, args...); out != scanHelp {
			t.Errorf("%v: output differs from scan -h", args)
		}
	}

	_, tuiHelp, _ := runCapturing(t, "tui", "-h")
	for _, args := range [][]string{{"tui", "--help"}, {"tui", "help"}, {"interactive", "-h"}, {"help", "tui"}} {
		if _, out, _ := runCapturing(t, args...); out != tuiHelp {
			t.Errorf("%v: output differs from tui -h", args)
		}
	}
}

func TestVersionOutput(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "1.2.3"
	for _, args := range [][]string{{"version"}, {"-version"}, {"--version"}, {"scan", "version"}, {"scan", "--version"}, {"tui", "version"}} {
		if _, stdout, _ := runCapturing(t, args...); stdout != "datasetlint 1.2.3\n" {
			t.Errorf("%v: version output = %q", args, stdout)
		}
	}

	Version = "dev"
	if _, stdout, _ := runCapturing(t, "--version"); !strings.HasPrefix(stdout, "datasetlint ") || strings.TrimSpace(stdout) == "datasetlint" {
		t.Errorf("dev version output = %q", stdout)
	}
}

func TestUnknownSubcommandPrintsUsageToStderr(t *testing.T) {
	for _, args := range [][]string{{"bogus"}, {"help", "bogus"}, {"-x"}} {
		code, stdout, stderr := runCapturing(t, args...)
		if code != 2 {
			t.Errorf("%v: exit code = %d, want 2", args, code)
		}
		if stdout != "" {
			t.Errorf("%v: unexpected stdout %q", args, stdout)
		}
		if !strings.Contains(stderr, "unknown subcommand") || !strings.Contains(stderr, "Usage:") {
			t.Errorf("%v: stderr should hold the error and usage, got %q", args, stderr)
		}
	}
}

func TestRealScanErrorsKeepExitCodes(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.jsonl")
	data := filepath.Join(dir, "data.jsonl")
	if err := os.WriteFile(data, []byte(`{"id":"1","input":"hello world","output":"hi"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dupes := filepath.Join(dir, "dupes.jsonl")
	if err := os.WriteFile(dupes, []byte(`{"id":"1","input":"same","output":"a"}`+"\n"+`{"id":"2","input":"same","output":"b"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		args       []string
		want       int
		wantStderr bool
	}{
		{[]string{"scan"}, 2, true},
		{[]string{"scan", "-train", data}, 2, true},
		{[]string{"scan", "-eval", data}, 2, true},
		{[]string{"scan", "-bogus"}, 2, true},
		{[]string{"scan", "-max-samples", "abc"}, 2, true},
		{[]string{"scan", "-train", missing, "-eval", data}, 1, true},
		{[]string{"scan", "-train", data, "-eval", missing}, 1, true},
		{[]string{"scan", "-train", data, "-eval", data}, 0, false},
		{[]string{"scan", "-train", data, "-eval", data, "-strict"}, 1, false},
		{[]string{"scan", "-train", dupes, "-eval", data}, 0, false},
		{[]string{"scan", "-train", dupes, "-eval", data, "-strict"}, 1, false},
		{[]string{"scan", "-train", dupes, "-eval", data, "-strict", "-json"}, 1, false},
	}
	for _, tc := range cases {
		code, _, stderr := runCapturing(t, tc.args...)
		if code != tc.want {
			t.Errorf("%v: exit code = %d, want %d", tc.args, code, tc.want)
		}
		if tc.wantStderr && stderr == "" {
			t.Errorf("%v: expected an error message on stderr", tc.args)
		}
		if !tc.wantStderr && stderr != "" {
			t.Errorf("%v: unexpected stderr %q", tc.args, stderr)
		}
	}
}

func TestFlagValueNamedHelpOrVersionIsNotHijacked(t *testing.T) {
	dir := t.TempDir()
	eval := filepath.Join(dir, "eval.jsonl")
	if err := os.WriteFile(eval, []byte(`{"id":"1","input":"hello","output":"hi"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// "help"/"version" here are values of -train, not subcommand words, so
	// they are treated as (missing) file paths: a failure (1), never help (0).
	for _, value := range []string{"help", "version", "--version"} {
		code, stdout, stderr := runCapturing(t, "scan", "-train", value, "-eval", eval)
		if code != 1 {
			t.Errorf("-train %s: exit code = %d, want 1", value, code)
		}
		if stdout != "" || !strings.Contains(stderr, value) {
			t.Errorf("-train %s: stdout %q stderr %q, want the file error on stderr", value, stdout, stderr)
		}
	}

	// A real file named "help" is read as data.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.WriteFile("help", []byte(`{"id":"1","input":"hello","output":"hi"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, _ := runCapturing(t, "scan", "-train", "help", "-eval", eval)
	if code != 0 || !strings.Contains(stdout, "Train rows: 1") {
		t.Errorf("-train help (existing file): code %d, stdout %q", code, stdout)
	}
}
