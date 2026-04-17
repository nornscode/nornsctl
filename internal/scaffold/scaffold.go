package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/amackera/nornsctl/internal/dev"
)

// SupportedLanguages returns the list of available template languages.
func SupportedLanguages() []string {
	return []string{"python"}
}

// IsSupported checks if a language has templates available.
func IsSupported(language string) bool {
	for _, l := range SupportedLanguages() {
		if l == language {
			return true
		}
	}
	return false
}

// Config holds the template variables and scaffold options.
type Config struct {
	Name        string
	PackageName string
	NornsURL    string
	NornsAPIKey string
	Language    string
	OutputDir   string
}

// Run scaffolds a new agent project.
func Run(cfg Config) error {
	// Check target dir
	if err := checkOutputDir(cfg.OutputDir); err != nil {
		return err
	}

	// Walk the embedded template tree for the language
	root := filepath.Join("templates", cfg.Language)
	err := fs.WalkDir(templates, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path from template root
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}

		// Rewrite path segments: replace {{.PackageName}} with actual name
		outRel := rewritePath(rel, cfg.PackageName)

		outPath := filepath.Join(cfg.OutputDir, outRel)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		// Read source
		data, err := templates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		// If .tmpl, render through text/template and strip extension
		if strings.HasSuffix(outPath, ".tmpl") {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
			content, err := renderTemplate(string(data), cfg)
			if err != nil {
				return fmt.Errorf("rendering %s: %w", path, err)
			}
			return os.WriteFile(outPath, []byte(content), 0644)
		}

		// Otherwise copy verbatim
		return os.WriteFile(outPath, data, 0644)
	})
	if err != nil {
		return err
	}

	// Auto-wire .env from dev state
	devWired := false
	state := loadDevState()
	if state != nil {
		envContent := fmt.Sprintf("NORNS_URL=%s\nNORNS_API_KEY=%s\nANTHROPIC_API_KEY=\n", state.URL, state.APIKey)
		envPath := filepath.Join(cfg.OutputDir, ".env")
		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("writing .env: %w", err)
		}
		devWired = true
	}

	// Print success
	absPath, _ := filepath.Abs(cfg.OutputDir)
	fmt.Printf("\nCreated agent project %q at %s\n", cfg.Name, absPath)

	if devWired {
		fmt.Printf("\n  Auto-configured from dev server (%s)\n", state.URL)
		fmt.Printf("\n  Next steps:\n")
		fmt.Printf("    cd %s\n", cfg.OutputDir)
		fmt.Printf("    uv sync\n")
		fmt.Printf("    echo 'ANTHROPIC_API_KEY=sk-ant-...' >> .env\n")
		fmt.Printf("    uv run %s-worker\n\n", cfg.Name)
		fmt.Printf("  Then in another terminal, send a test message:\n")
		fmt.Printf("    cd %s\n", cfg.OutputDir)
		fmt.Printf("    uv run %s-client\n", cfg.Name)
	} else {
		fmt.Printf("\n  Tip: run `nornsctl dev` to start a local server and auto-configure\n")
		fmt.Printf("\n  Next steps:\n")
		fmt.Printf("    cd %s\n", cfg.OutputDir)
		fmt.Printf("    uv sync\n")
		fmt.Printf("    cp .env.example .env   # fill in your API keys\n")
		fmt.Printf("    uv run %s-worker\n\n", cfg.Name)
		fmt.Printf("  Then in another terminal, send a test message:\n")
		fmt.Printf("    cd %s\n", cfg.OutputDir)
		fmt.Printf("    uv run %s-client\n", cfg.Name)
	}
	fmt.Println()

	return nil
}

func checkOutputDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0755)
		}
		return err
	}
	if len(entries) > 0 {
		absPath, _ := filepath.Abs(dir)
		return fmt.Errorf("refusing to scaffold into non-empty directory: %s", absPath)
	}
	return nil
}

func rewritePath(path, packageName string) string {
	parts := strings.Split(path, string(filepath.Separator))
	for i, part := range parts {
		if part == "{{.PackageName}}" {
			parts[i] = packageName
		}
	}
	return filepath.Join(parts...)
}

func renderTemplate(content string, cfg Config) (string, error) {
	tmpl, err := template.New("").Parse(content)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func loadDevState() *dev.State {
	state, err := dev.LoadState()
	if err != nil || state == nil {
		return nil
	}
	if state.URL == "" || state.APIKey == "" {
		return nil
	}
	return state
}
