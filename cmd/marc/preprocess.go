package main

import "strings"

// PreprocessNewpage rewrites literal `\newpage` markers in Markdown
// source into pandoc's `::: pagebreak :::` fenced-div syntax, matching
// the sed step from the original render.sh script. `\newpage` is not
// native Markdown/pandoc syntax; the pagebreak div, handled via
// template.html's raw-attribute CSS, is.
func PreprocessNewpage(src string) string {
	return strings.ReplaceAll(src, `\newpage`, "::: pagebreak\n:::")
}

// ExtractTitle returns the text of the first top-level Markdown
// heading (a line beginning with "# ") in src, or "" if none is
// found. Mirrors render.sh's `grep '^# ' | head -1` step, used to set
// the rendered document's <title>.
func ExtractTitle(src string) string {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}
