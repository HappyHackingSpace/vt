package template

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// GetFiltered returns templates matching the given filter and tags, sorted by ID.
// filter is matched as a case-insensitive substring against the template ID and
// name. tags are matched exactly (case-insensitive, whitespace-trimmed) against
// template tags with OR logic across the given tags. When both filter and tags
// are provided, a template must satisfy both conditions.
func GetFiltered(templates map[string]Template, filter string, tags []string) []Template {
	normalizedFilter := strings.ToLower(strings.TrimSpace(filter))
	normalizedTags := normalizeTags(tags)

	matched := make([]Template, 0, len(templates))
	for _, tmpl := range templates {
		if normalizedFilter != "" && !matchesFilter(tmpl, normalizedFilter) {
			continue
		}
		if len(normalizedTags) > 0 && !matchesAnyTag(tmpl, normalizedTags) {
			continue
		}
		matched = append(matched, tmpl)
	}

	slices.SortFunc(matched, func(a, b Template) int {
		return strings.Compare(a.ID, b.ID)
	})

	return matched
}

// ListFilteredTemplates displays templates matching the given filter and tags
// in a table format, sorted by ID.
func ListFilteredTemplates(templates map[string]Template, filter string, tags []string) {
	matched := GetFiltered(templates, filter, tags)

	if len(matched) == 0 {
		if strings.TrimSpace(filter) != "" || len(normalizeTags(tags)) > 0 {
			fmt.Println("No templates found matching the given filters")
		} else {
			fmt.Println("No templates found")
		}
		return
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Author", "Targets", "Type", "Tags"})

	for _, tmpl := range matched {
		t.AppendRow(table.Row{
			tmpl.ID,
			tmpl.Info.Name,
			tmpl.Info.Author,
			strings.Join(tmpl.Info.Targets, ", "),
			tmpl.Info.Type,
			strings.Join(tmpl.Info.Tags, ", "),
		})
	}

	t.SetCaption("there are %d templates", len(matched))
	t.SetIndexColumn(0)
	t.Render()
}

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func matchesFilter(tmpl Template, normalizedFilter string) bool {
	return strings.Contains(strings.ToLower(tmpl.ID), normalizedFilter) ||
		strings.Contains(strings.ToLower(tmpl.Info.Name), normalizedFilter)
}

func matchesAnyTag(tmpl Template, normalizedTags []string) bool {
	for _, tag := range tmpl.Info.Tags {
		if slices.Contains(normalizedTags, strings.ToLower(strings.TrimSpace(tag))) {
			return true
		}
	}
	return false
}
