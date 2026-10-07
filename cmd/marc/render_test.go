package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPandocInputFormat_listDirectlyAfterParagraph(t *testing.T) {
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		t.Skip("pandoc not on PATH")
	}

	cmd := exec.Command(pandoc, "--from", pandocInputFormat, "--to", "html")
	cmd.Stdin = strings.NewReader("Intro:\n- one\n- two\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pandoc failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "<li>one</li>") {
		t.Errorf("list was not rendered as a list, got:\n%s", out)
	}
}
