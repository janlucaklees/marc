package main

import "fmt"

// ResolveBinary determines which executable path to use for a tool.
// If configured is non-empty, it is checked via lookup and returned
// (or an error if lookup fails). Otherwise each name in candidates is
// tried via lookup, in order; the first that resolves is returned. If
// nothing resolves, an error names toolLabel and every candidate
// tried.
func ResolveBinary(toolLabel, configured string, candidates []string, lookup func(string) (string, error)) (string, error) {
	if configured != "" {
		path, err := lookup(configured)
		if err != nil {
			return "", fmt.Errorf("configured %s binary %q not usable: %w", toolLabel, configured, err)
		}
		return path, nil
	}
	var tried []string
	for _, name := range candidates {
		if path, err := lookup(name); err == nil {
			return path, nil
		}
		tried = append(tried, name)
	}
	return "", fmt.Errorf("no %s binary found (tried: %v); set %s_bin in config.toml or install one on PATH", toolLabel, tried, toolLabel)
}
