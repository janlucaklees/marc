package main

import "testing"

func TestDefaultOutputPath_replacesExtension(t *testing.T) {
	got := DefaultOutputPath("/home/jlk/docs/Readme.md")
	want := "/home/jlk/docs/Readme.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}

func TestDefaultOutputPath_relativePath(t *testing.T) {
	got := DefaultOutputPath("notes.markdown")
	want := "notes.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}

func TestDefaultOutputPath_noExtension(t *testing.T) {
	got := DefaultOutputPath("README")
	want := "README.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}
