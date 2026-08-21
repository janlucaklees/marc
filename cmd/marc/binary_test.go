package main

import (
	"errors"
	"testing"
)

func TestResolveBinary_usesConfiguredWhenSet(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "/custom/pandoc" {
			return "/custom/pandoc", nil
		}
		return "", errors.New("not found")
	}
	path, err := ResolveBinary("pandoc", "/custom/pandoc", []string{"pandoc"}, lookup)
	if err != nil {
		t.Fatalf("ResolveBinary() error = %v", err)
	}
	if path != "/custom/pandoc" {
		t.Errorf("ResolveBinary() = %q, want %q", path, "/custom/pandoc")
	}
}

func TestResolveBinary_configuredButUnusableReturnsError(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("no such file")
	}
	_, err := ResolveBinary("pandoc", "/bad/path", []string{"pandoc"}, lookup)
	if err == nil {
		t.Error("ResolveBinary() error = nil, want error for unusable configured path")
	}
}

func TestResolveBinary_triesCandidatesInOrder(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "chromium-browser" {
			return "/usr/bin/chromium-browser", nil
		}
		return "", errors.New("not found")
	}
	path, err := ResolveBinary("chromium", "", []string{"chromium", "chromium-browser", "google-chrome"}, lookup)
	if err != nil {
		t.Fatalf("ResolveBinary() error = %v", err)
	}
	if path != "/usr/bin/chromium-browser" {
		t.Errorf("ResolveBinary() = %q, want %q", path, "/usr/bin/chromium-browser")
	}
}

func TestResolveBinary_noneFoundReturnsError(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	_, err := ResolveBinary("chromium", "", []string{"chromium", "chromium-browser"}, lookup)
	if err == nil {
		t.Error("ResolveBinary() error = nil, want error when no candidate resolves")
	}
}
