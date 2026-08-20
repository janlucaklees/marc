package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var chromiumCandidates = []string{
	"chromium",
	"chromium-browser",
	"google-chrome",
	"google-chrome-stable",
	"brave-browser",
}

func main() {
	os.Exit(run())
}

func run() int {
	var templateName string
	var outputPath string

	fs := flag.NewFlagSet("marc", flag.ContinueOnError)
	fs.StringVar(&templateName, "template", "", "template name (dotfiles-managed)")
	fs.StringVar(&templateName, "t", "", "shorthand for -template")
	fs.StringVar(&outputPath, "output", "", "output PDF path (default: input basename with .pdf)")
	fs.StringVar(&outputPath, "o", "", "shorthand for -output")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: marc [-t|--template <name>] [-o|--output <path>] <input.md>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return 2
	}

	if fs.NArg() != 1 {
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
	chromiumBin, err := ResolveBinary("chromium", cfg.ChromiumBin, chromiumCandidates, exec.LookPath)
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
			fmt.Fprintf(os.Stderr, "marc: no templates found in %s\n", templatesDir)
			return 1
		}
		if !IsInteractive(os.Stdin) {
			fmt.Fprintf(os.Stderr, "marc: %v\n", ErrNoTemplate)
			return 1
		}
		templateName, err = PromptTemplate(names, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marc: %v\n", err)
			return 1
		}
	}

	templatePath, err := TemplatePath(templatesDir, templateName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
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
