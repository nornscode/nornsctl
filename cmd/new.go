package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/amackera/nornsctl/internal/scaffold"
	"github.com/spf13/cobra"
)

var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)
var singleCharRegex = regexp.MustCompile(`^[a-z]$`)

var pythonReserved = map[string]bool{
	"async": true, "class": true, "import": true, "type": true,
	"def": true, "return": true, "yield": true, "lambda": true,
	"from": true, "global": true, "nonlocal": true, "pass": true,
	"break": true, "continue": true, "if": true, "else": true,
	"elif": true, "for": true, "while": true, "try": true,
	"except": true, "finally": true, "with": true, "as": true,
	"is": true, "in": true, "not": true, "and": true, "or": true,
	"del": true, "raise": true, "assert": true,
	"true": true, "false": true, "none": true,
}

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Scaffold a new agent project",
	Long:  "Create a new Norns agent worker project from a template.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := validateName(name); err != nil {
			return err
		}

		language, _ := cmd.Flags().GetString("language")
		if !scaffold.IsSupported(language) {
			return fmt.Errorf("unsupported language: %s. Supported: %s", language, strings.Join(scaffold.SupportedLanguages(), ", "))
		}

		tmpl, _ := cmd.Flags().GetString("template")
		if !scaffold.IsSupportedTemplate(language, tmpl) {
			return fmt.Errorf("unknown template: %s. Available for %s: %s", tmpl, language, strings.Join(scaffold.SupportedTemplates(language), ", "))
		}

		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = name
		}

		packageName := strings.ReplaceAll(name, "-", "_")

		return scaffold.Run(scaffold.Config{
			Name:        name,
			PackageName: packageName,
			Language:    language,
			Template:    tmpl,
			OutputDir:   dir,
		})
	},
}

func validateName(name string) error {
	if len(name) > 64 {
		return fmt.Errorf("project name must be 64 characters or fewer")
	}

	if !singleCharRegex.MatchString(name) && !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must be lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen", name)
	}

	if strings.Contains(name, "--") {
		return fmt.Errorf("invalid project name %q: consecutive hyphens are not allowed", name)
	}

	packageName := strings.ReplaceAll(name, "-", "_")
	if pythonReserved[packageName] {
		return fmt.Errorf("invalid project name %q: %q is a Python reserved word", name, packageName)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().String("language", "python", "Template language (supported: python)")
	newCmd.Flags().String("template", "default", "Project template (default, slack-bot)")
	newCmd.Flags().String("dir", "", "Output directory (default: ./<name>)")
}
