package main

import (
	"reflect"
	"testing"
)

func TestChromiumCandidates_linuxIsPathNames(t *testing.T) {
	want := []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "brave-browser"}
	if got := chromiumCandidates("linux"); !reflect.DeepEqual(got, want) {
		t.Errorf("chromiumCandidates(linux) = %v, want %v", got, want)
	}
}

func TestChromiumCandidates_unknownOSFallsBackToPathNames(t *testing.T) {
	if got, want := chromiumCandidates("freebsd"), chromiumCandidates("linux"); !reflect.DeepEqual(got, want) {
		t.Errorf("chromiumCandidates(freebsd) = %v, want linux list %v", got, want)
	}
}

func TestChromiumCandidates_darwinPrefersAppBundlesThenPath(t *testing.T) {
	got := chromiumCandidates("darwin")

	first := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if got[0] != first {
		t.Errorf("chromiumCandidates(darwin)[0] = %q, want %q", got[0], first)
	}
	if last := got[len(got)-1]; last != "brave-browser" {
		t.Errorf("chromiumCandidates(darwin) should end with PATH names, last = %q", last)
	}
}
