package main

import (
	"strings"
	"testing"
)

const sampleHTML = "<html><head><title>t</title></head><body><main>x</main></body></html>"

func TestUnpagedHTMLInjectsPageSizeWithSlack(t *testing.T) {
	got, err := UnpagedHTML(sampleHTML, 1000)
	if err != nil {
		t.Fatal(err)
	}
	want := "@page { size: 900px 1004px; margin: 0; }"
	if !strings.Contains(got, want) {
		t.Errorf("missing %q in:\n%s", want, got)
	}
	if strings.Index(got, "@page") > strings.Index(got, "</head>") {
		t.Error("style must be injected inside <head>")
	}
}

func TestUnpagedHTMLRequiresHead(t *testing.T) {
	if _, err := UnpagedHTML("<body></body>", 100); err == nil {
		t.Error("expected error for HTML without </head>")
	}
}

func TestMeasurementHTMLInjectsStyleAndScript(t *testing.T) {
	got, err := MeasurementHTML(sampleHTML)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "data-marc-height") {
		t.Error("missing measuring script")
	}
	if strings.Index(got, "<script>") < strings.Index(got, "</main>") {
		t.Error("script must come after the body content")
	}
	if strings.Contains(got, "@page") {
		t.Error("measurement pass must not set a page size")
	}
}

func TestParseMeasuredHeight(t *testing.T) {
	h, err := ParseMeasuredHeight(`<html><body data-marc-height="2345"><main></main></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if h != 2345 {
		t.Errorf("got %d, want 2345", h)
	}
}

func TestParseMeasuredHeightMissing(t *testing.T) {
	if _, err := ParseMeasuredHeight("<html><body></body></html>"); err == nil {
		t.Error("expected error when height attribute is absent")
	}
}
