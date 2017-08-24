package main

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestMainCommands(t *testing.T) {
	testCommands := []string{
		"init",
		"apps",
		"routes",
		"images",
		"lambda",
		"version",
		"build",
		"bump",
		"deploy",
		"run",
		"push",
		"logs",
		"calls",
		"call",
	}
	tmp := os.TempDir()
	fnTestBin := path.Join(tmp, "fn-test")
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	err := exec.Command("go", "build", "-o", fnTestBin).Run()
	if err != nil {
		t.Fatalf("Failed to build fn: err: %s", err)
	}

	for _, cmd := range testCommands {
		res, err := exec.Command(fnTestBin, strings.Split(cmd, " ")...).CombinedOutput()
		if bytes.Contains(res, []byte("command not found")) {
			t.Error(err)
		}
	}

	os.Remove(fnTestBin)
}
