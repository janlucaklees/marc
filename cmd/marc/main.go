package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	os.Exit(run())
}

func run() int {
	var templateName string
	var outputPath string
	var interactive bool

	fs := flag.NewFlagSet("marc", flag.ContinueOnError)
	fs.StringVar(&templateName, "template", "", "template name (dotfiles-managed)")
	fs.StringVar(&templateName, "t", "", "shorthand for -template")
	fs.StringVar(&outputPath, "output", "", "output PDF path (default: input basename with .pdf)")
	fs.StringVar(&outputPath, "o", "", "shorthand for -output")
	fs.BoolVar(&interactive, "interactive", false, "always prompt for a template, ignoring config's default_template")
	fs.BoolVar(&interactive, "i", false, "shorthand for -interactive")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: marc [-t|--template <name>] [-i|--interactive] [-o|--output <path>] <input.md>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(permuteArgs(os.Args[1:])); err != nil {
		return 2
	}

	if fs.NArg() != 1 {
		if fs.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "marc: missing input file argument")
		} else {
			fmt.Fprintf(os.Stderr, "marc: unexpected arguments: %v\n", fs.Args())
		}
		fs.Usage()
		return 2
	}
	inputPath := fs.Arg(0)

	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	configDir, err := ConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: resolving config directory: %v\n", err)
		return 1
	}

	cfg, err := LoadConfig(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: loading config: %v\n", err)
		return 1
	}

	pandocBin, err := ResolveBinary("pandoc", cfg.PandocBin, []string{"pandoc"}, exec.LookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}
	chromiumBin, err := ResolveBinary("chromium", cfg.ChromiumBin, chromiumCandidates(runtime.GOOS), exec.LookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	templatesDir := TemplatesDir(configDir)
	if templateName == "" {
		names, err := ListTemplates(templatesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marc: listing templates: %v\n", err)
			return 1
		}
		if len(names) == 0 {
			if _, err := BootstrapDefaultTemplate(templatesDir); err != nil {
				fmt.Fprintf(os.Stderr, "marc: creating starter template: %v\n", err)
				return 1
			}
			if _, err := BootstrapDefaultConfig(configDir); err != nil {
				fmt.Fprintf(os.Stderr, "marc: creating starter config: %v\n", err)
				return 1
			}
			fmt.Fprintf(os.Stderr, "marc: no templates found — created a starter template (%s) and config (%s); using it for this run\n",
				filepath.Join(templatesDir, "default", "template.html"), filepath.Join(configDir, "config.toml"))
			names = []string{"default"}
		}

		defaultTemplate := cfg.DefaultTemplate
		if !cfg.DefaultTemplateSet {
			defaultTemplate = "default"
		}

		switch {
		case interactive || defaultTemplate == "":
			if !IsInteractive(os.Stdin) {
				fmt.Fprintf(os.Stderr, "marc: %v\n", ErrNoTemplate)
				return 1
			}
			templateName, err = PromptTemplate(names, os.Stdin, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "marc: %v\n", err)
				return 1
			}
		default:
			templateName = defaultTemplate
		}
	}

	templatePath, err := TemplatePath(templatesDir, templateName)
	if err != nil {
		names, listErr := ListTemplates(templatesDir)
		switch {
		case listErr != nil:
			fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		case len(names) == 0:
			fmt.Fprintf(os.Stderr, "marc: %v (no templates available in %s)\n", err, templatesDir)
		default:
			fmt.Fprintf(os.Stderr, "marc: %v (available templates: %s)\n", err, strings.Join(names, ", "))
		}
		return 1
	}

	if outputPath == "" {
		outputPath = DefaultOutputPath(inputPath)
	}

	if err := Render(RenderOptions{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		TemplatePath: templatePath,
		PandocBin:    pandocBin,
		ChromiumBin:  chromiumBin,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	fmt.Printf("Rendered %s\n", outputPath)
	return 0
}

// permuteArgs reorders args so that every flag (and its value, if it
// takes one) comes before all positional arguments. Go's flag package
// stops parsing at the first non-flag argument, so without this,
// "marc input.md -t work" would silently treat "-t" and "work" as
// extra positional arguments instead of a flag. "--" terminates
// permutation early: everything after it is treated as positional,
// matching flag.Parse's own handling of "--".
func permuteArgs(args []string) []string {
	// Flags that consume the following argument as their value (their
	// value isn't attached via "="). Keyed by the flag's exact spelling,
	// since flag.NewFlagSet registers both the short and long forms.
	valueFlags := map[string]bool{
		"-t": true, "--template": true,
		"-o": true, "--output": true,
	}

	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if a == "-" || !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		flags = append(flags, a)
		if !strings.Contains(a, "=") && valueFlags[a] && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...)
}
