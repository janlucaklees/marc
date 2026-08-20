package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustWriteTemplate(t *testing.T, templatesDir, name string) {
	t.Helper()
	dir := filepath.Join(templatesDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed creating template dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("failed writing template.html: %v", err)
	}
}

func TestListTemplates_returnsSortedNames(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	mustWriteTemplate(t, dir, "private")
	names, err := ListTemplates(dir)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	want := []string{"private", "work"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("ListTemplates() = %v, want %v", names, want)
	}
}

func TestListTemplates_ignoresDirsWithoutTemplateHTML(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0o755); err != nil {
		t.Fatalf("failed creating empty dir: %v", err)
	}
	names, err := ListTemplates(dir)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(names) != 1 || names[0] != "work" {
		t.Errorf("ListTemplates() = %v, want [work]", names)
	}
}

func TestListTemplates_missingDirReturnsEmpty(t *testing.T) {
	names, err := ListTemplates(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("ListTemplates() = %v, want empty", names)
	}
}

func TestTemplatePath_found(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	path, err := TemplatePath(dir, "work")
	if err != nil {
		t.Fatalf("TemplatePath() error = %v", err)
	}
	want := filepath.Join(dir, "work", "template.html")
	if path != want {
		t.Errorf("TemplatePath() = %q, want %q", path, want)
	}
}

func TestTemplatePath_notFoundReturnsError(t *testing.T) {
	dir := t.TempDir()
	if _, err := TemplatePath(dir, "missing"); err == nil {
		t.Error("TemplatePath() error = nil, want error for missing template")
	}
}

func TestPromptTemplate_selectsByExactName(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("work\n")
	var out strings.Builder
	got, err := PromptTemplate(names, in, &out)
	if err != nil {
		t.Fatalf("PromptTemplate() error = %v", err)
	}
	if got != "work" {
		t.Errorf("PromptTemplate() = %q, want %q", got, "work")
	}
}

func TestPromptTemplate_selectsByNumber(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("2\n")
	var out strings.Builder
	got, err := PromptTemplate(names, in, &out)
	if err != nil {
		t.Fatalf("PromptTemplate() error = %v", err)
	}
	if got != "work" {
		t.Errorf("PromptTemplate() = %q, want %q", got, "work")
	}
}

func TestPromptTemplate_invalidSelectionReturnsError(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("bogus\n")
	var out strings.Builder
	if _, err := PromptTemplate(names, in, &out); err == nil {
		t.Error("PromptTemplate() error = nil, want error for invalid selection")
	}
}
