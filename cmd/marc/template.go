package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ErrNoTemplate indicates no template name was given and the session
// is non-interactive, so no prompt could be shown.
var ErrNoTemplate = fmt.Errorf("no template specified; pass -t/--template (session is non-interactive)")

// defaultTemplateHTML is the starter template shipped with marc,
// written out to the templates dir on first run when none exist yet.
// See BootstrapDefaultTemplate.
//
//go:embed template.html
var defaultTemplateHTML []byte

// BootstrapDefaultTemplate writes the embedded starter template to
// templatesDir/default/template.html if templatesDir has no templates
// at all yet. It reports whether it created one; a false, nil result
// means templates already exist and nothing was done.
func BootstrapDefaultTemplate(templatesDir string) (bool, error) {
	names, err := ListTemplates(templatesDir)
	if err != nil {
		return false, err
	}
	if len(names) > 0 {
		return false, nil
	}
	dir := filepath.Join(templatesDir, "default")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(dir, "template.html"), defaultTemplateHTML, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// TemplatesDir returns the templates directory beneath configDir.
func TemplatesDir(configDir string) string {
	return filepath.Join(configDir, "templates")
}

// ListTemplates returns the names of available templates in
// templatesDir: subdirectories that contain a template.html file,
// sorted alphabetically. If templatesDir does not exist, it returns
// an empty slice and no error.
func ListTemplates(templatesDir string) ([]string, error) {
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(templatesDir, e.Name(), "template.html")); err == nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// TemplatePath returns the template.html path for name within
// templatesDir. Returns an error if that file doesn't exist.
func TemplatePath(templatesDir, name string) (string, error) {
	path := filepath.Join(templatesDir, name, "template.html")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("template %q not found (looked for %s)", name, path)
	}
	return path, nil
}

// IsInteractive reports whether f looks like a terminal, as opposed
// to a pipe or redirected file.
func IsInteractive(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// PromptTemplate writes a numbered list of names to out, reads a
// single line from in, and returns the selected name. The line may be
// either the exact template name or its 1-based list number.
func PromptTemplate(names []string, in io.Reader, out io.Writer) (string, error) {
	fmt.Fprintln(out, "Available templates:")
	for i, n := range names {
		fmt.Fprintf(out, "  %d) %s\n", i+1, n)
	}
	fmt.Fprint(out, "Select a template: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return "", fmt.Errorf("no template selection received")
	}
	choice := strings.TrimSpace(scanner.Text())

	for _, n := range names {
		if n == choice {
			return n, nil
		}
	}
	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(names) {
		return names[idx-1], nil
	}
	return "", fmt.Errorf("invalid template selection %q", choice)
}
