package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Unpaged rendering produces a PDF with a single page that is exactly
// as tall as the document. Chromium's --print-to-pdf has no page-size
// flag, but it honors a CSS @page size, so Render measures the content
// height in a first headless pass and bakes it into @page for the
// second (print) pass.
const (
	// unpagedWidthPx is both the Chromium viewport width during
	// measurement and the PDF page width, so the layout is identical in
	// both passes.
	unpagedWidthPx = 900

	// unpagedSlackPx keeps sub-pixel rounding from spilling onto a
	// second page.
	unpagedSlackPx = 4

	// unpagedBaseCSS pins the on-screen look in the print pass too. It
	// is injected as plain (non-@media) CSS after the template's styles,
	// so it wins over a template's own @media print overrides. The
	// values mirror the starter template's screen layout.
	unpagedBaseCSS = `html { width: 900px; }
body { padding: 3rem 1.5rem; font-size: 16px; }
main { max-width: 740px; }`

	// measureScript stores the laid-out document height on <body>
	// shortly after load, giving async content (Mermaid) time to render.
	measureScript = `<script>
window.addEventListener('load', function () {
  setTimeout(function () {
    document.body.setAttribute('data-marc-height', String(Math.ceil(document.documentElement.scrollHeight)));
  }, 1000);
});
</script>`
)

var measuredHeightRe = regexp.MustCompile(`data-marc-height="(\d+)"`)

// MeasurementHTML returns html prepared for the measuring pass: the
// unpaged base CSS plus the script that records the document height.
func MeasurementHTML(html string) (string, error) {
	withCSS, err := insertBefore(html, "</head>", "<style>\n"+unpagedBaseCSS+"\n</style>\n")
	if err != nil {
		return "", err
	}
	return insertBefore(withCSS, "</body>", measureScript+"\n")
}

// UnpagedHTML returns html with the unpaged base CSS and a single
// @page sized to heightPx (plus slack) injected.
func UnpagedHTML(html string, heightPx int) (string, error) {
	css := fmt.Sprintf("<style>\n%s\n@page { size: %dpx %dpx; margin: 0; }\n</style>\n",
		unpagedBaseCSS, unpagedWidthPx, heightPx+unpagedSlackPx)
	return insertBefore(html, "</head>", css)
}

// ParseMeasuredHeight extracts the height recorded by measureScript
// from the DOM that Chromium's --dump-dom printed.
func ParseMeasuredHeight(dom string) (int, error) {
	m := measuredHeightRe.FindStringSubmatch(dom)
	if m == nil {
		return 0, fmt.Errorf("document height not found in rendered DOM (did the page finish loading?)")
	}
	return strconv.Atoi(m[1])
}

func insertBefore(html, tag, insert string) (string, error) {
	i := strings.Index(html, tag)
	if i < 0 {
		return "", fmt.Errorf("template output has no %s tag; cannot render unpaged", tag)
	}
	return html[:i] + insert + html[i:], nil
}
