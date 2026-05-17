package template

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

// Template represents a vulnerable target environment configuration.
type Template struct {
	ID             string                    `yaml:"id"`
	Info           Info                      `yaml:"info"`
	ProofOfConcept map[string]any            `yaml:"poc"`
	Remediation    []string                  `yaml:"remediation"`
	Providers      map[string]ProviderConfig `yaml:"providers"`
	PostInstall    []string                  `yaml:"post-install"`
}

// Info contains metadata about a template.
type Info struct {
	Name             string   `yaml:"name"`
	Description      string   `yaml:"description"`
	Author           string   `yaml:"author"`
	Targets          []string `yaml:"targets"`
	Type             string   `yaml:"type"`
	AffectedVersions []string `yaml:"affected_versions"`
	FixedVersion     string   `yaml:"fixed_version"`
	Cwe              string   `yaml:"cwe"`
	Cvss             Cvss     `yaml:"cvss"`
	Tags             []string `yaml:"tags"`
	References       []string `yaml:"references"`
}

// ProviderConfig contains configuration for a specific provider.
type ProviderConfig struct {
	Path  string `yaml:"path"`
	Image string `yaml:"image,omitempty"`
	Ports []int  `yaml:"ports,omitempty"`
}

// Cvss represents Common Vulnerability Scoring System information.
type Cvss struct {
	Score   string `yaml:"score"`
	Metrics string `yaml:"metrics"`
}

// String returns template fields as a table.
func (t Template) String() string {
	tw := table.NewWriter()
	tw.AppendRow(table.Row{"ID", t.ID})
	tw.AppendRow(table.Row{"Name", t.Info.Name})
	tw.AppendRow(table.Row{"Description", t.Info.Description})
	tw.AppendRow(table.Row{"Author", t.Info.Author})
	tw.AppendRow(table.Row{"Type", t.Info.Type})
	tw.AppendRow(table.Row{"Targets", formatList(t.Info.Targets)})
	tw.AppendRow(table.Row{"Affected Versions", formatList(t.Info.AffectedVersions)})
	tw.AppendRow(table.Row{"Fixed Version", t.Info.FixedVersion})
	tw.AppendRow(table.Row{"CWE", t.Info.Cwe})
	tw.AppendRow(table.Row{"CVSS Score", t.Info.Cvss.Score})
	tw.AppendRow(table.Row{"CVSS Metrics", t.Info.Cvss.Metrics})
	tw.AppendRow(table.Row{"Tags", formatList(t.Info.Tags)})
	tw.AppendRow(table.Row{"References", formatList(t.Info.References)})
	tw.AppendRow(table.Row{"Proof of Concept", formatPoc(t.ProofOfConcept)})
	tw.AppendRow(table.Row{"Remediation", formatList(t.Remediation)})
	tw.AppendRow(table.Row{"Providers", formatProviders(t.Providers)})
	tw.AppendRow(table.Row{"Post Install", formatList(t.PostInstall)})

	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateRows = true
	tw.Style().Options.SeparateColumns = true

	return tw.Render()
}

// ListTemplates displays all available templates in a table format.
func ListTemplates(templates map[string]Template) {
	ListTemplatesWithFilter(templates, "")
}

// ListTemplatesWithFilter displays templates in a table format, optionally filtered by tag.
func ListTemplatesWithFilter(templates map[string]Template, filterTag string) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Author", "Targets", "Type", "Tags"})

	count := 0
	for _, tmpl := range templates {
		if filterTag != "" {
			hasTag := false
			for _, tag := range tmpl.Info.Tags {
				if strings.EqualFold(tag, filterTag) || strings.Contains(strings.ToLower(tag), strings.ToLower(filterTag)) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		t.AppendRow(table.Row{
			tmpl.ID,
			tmpl.Info.Name,
			tmpl.Info.Author,
			strings.Join(tmpl.Info.Targets, ", "),
			tmpl.Info.Type,
			strings.Join(tmpl.Info.Tags, ", "),
		})
		count++
	}

	if count == 0 {
		if filterTag != "" {
			fmt.Printf("No templates found with tag matching '%s'\n", filterTag)
		} else {
			fmt.Println("No templates found")
		}
		return
	}

	if filterTag != "" {
		t.SetCaption("Found %d templates with tag matching '%s'", count, filterTag)
	} else {
		t.SetCaption("there are %d templates", count)
	}
	t.SetIndexColumn(0)
	t.Render()
}

// GetByID retrieves a template by its ID from the given templates map.
func GetByID(templates map[string]Template, templateID string) (*Template, error) {
	tmpl, ok := templates[templateID]
	if !ok || tmpl.ID == "" {
		return nil, fmt.Errorf("template %s not found", templateID)
	}
	return &tmpl, nil
}

func formatPoc(poc map[string]any) string {
	if len(poc) == 0 {
		return ""
	}
	var parts []string
	for key, value := range poc {
		parts = append(parts, fmt.Sprintf("%s: %s", key, formatPocValue(value)))
	}
	return strings.Join(parts, "\n")
}

func formatPocValue(v any) string {
	switch val := v.(type) {
	case []any:
		strs := make([]string, 0, len(val))
		for _, item := range val {
			strs = append(strs, fmt.Sprintf("%v", item))
		}
		return strings.Join(strs, ", ")
	case []string:
		return strings.Join(val, ", ")
	case map[string]any:
		var nested []string
		for k, nv := range val {
			nested = append(nested, fmt.Sprintf("%s=%v", k, nv))
		}
		return "{" + strings.Join(nested, ", ") + "}"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatProviders(providers map[string]ProviderConfig) string {
	if len(providers) == 0 {
		return ""
	}
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return strings.Join(names, "\n")
}

func formatList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, "\n")
}
