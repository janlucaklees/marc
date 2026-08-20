package main

import "testing"

func TestPreprocessNewpage_replacesMarker(t *testing.T) {
	input := "Section one.\n\n\\newpage\n\n## Section two\n"
	want := "Section one.\n\n::: pagebreak\n:::\n\n## Section two\n"
	got := PreprocessNewpage(input)
	if got != want {
		t.Errorf("PreprocessNewpage() = %q, want %q", got, want)
	}
}

func TestPreprocessNewpage_multipleMarkers(t *testing.T) {
	input := "a\\newpageb\\newpagec"
	want := "a::: pagebreak\n:::b::: pagebreak\n:::c"
	got := PreprocessNewpage(input)
	if got != want {
		t.Errorf("PreprocessNewpage() = %q, want %q", got, want)
	}
}

func TestPreprocessNewpage_noMarker(t *testing.T) {
	input := "nothing to replace here"
	got := PreprocessNewpage(input)
	if got != input {
		t.Errorf("PreprocessNewpage() = %q, want unchanged %q", got, input)
	}
}

func TestExtractTitle_findsFirstHeading(t *testing.T) {
	src := "intro text\n# Systemdokumentation Renaissance\n\nmore text\n# Second Heading\n"
	got := ExtractTitle(src)
	want := "Systemdokumentation Renaissance"
	if got != want {
		t.Errorf("ExtractTitle() = %q, want %q", got, want)
	}
}

func TestExtractTitle_noHeadingReturnsEmpty(t *testing.T) {
	src := "no headings here\njust text\n"
	got := ExtractTitle(src)
	if got != "" {
		t.Errorf("ExtractTitle() = %q, want empty string", got)
	}
}

func TestExtractTitle_ignoresSubheadings(t *testing.T) {
	src := "## Not a title\n# Actual Title\n"
	got := ExtractTitle(src)
	want := "Actual Title"
	if got != want {
		t.Errorf("ExtractTitle() = %q, want %q", got, want)
	}
}
