package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFiltered(t *testing.T) {
	templates := map[string]Template{
		"vt-dvwa": {
			ID: "vt-dvwa",
			Info: Info{
				Name: "Damn Vulnerable Web Application",
				Tags: []string{"sqli", "xss", "web"},
			},
		},
		"vt-juice-shop": {
			ID: "vt-juice-shop",
			Info: Info{
				Name: "OWASP Juice Shop",
				Tags: []string{"owasp", "web"},
			},
		},
		"vt-webgoat": {
			ID: "vt-webgoat",
			Info: Info{
				Name: "WebGoat",
				Tags: []string{"OWASP", "api"},
			},
		},
	}

	tests := []struct {
		name    string
		filter  string
		tags    []string
		wantIDs []string
	}{
		{
			name:    "no filters returns all sorted by ID",
			wantIDs: []string{"vt-dvwa", "vt-juice-shop", "vt-webgoat"},
		},
		{
			name:    "filter matches ID substring",
			filter:  "dvwa",
			wantIDs: []string{"vt-dvwa"},
		},
		{
			name:    "filter matches name substring",
			filter:  "juice",
			wantIDs: []string{"vt-juice-shop"},
		},
		{
			name:    "filter is case-insensitive",
			filter:  "WEBGOAT",
			wantIDs: []string{"vt-webgoat"},
		},
		{
			name:    "filter with no match returns empty",
			filter:  "does-not-exist",
			wantIDs: []string{},
		},
		{
			name:    "tag matches exactly",
			tags:    []string{"sqli"},
			wantIDs: []string{"vt-dvwa"},
		},
		{
			name:    "tag match is case-insensitive",
			tags:    []string{"OWASP"},
			wantIDs: []string{"vt-juice-shop", "vt-webgoat"},
		},
		{
			name:    "tag substring does not match",
			tags:    []string{"owa"},
			wantIDs: []string{},
		},
		{
			name:    "multiple tags use OR logic",
			tags:    []string{"sqli", "api"},
			wantIDs: []string{"vt-dvwa", "vt-webgoat"},
		},
		{
			name:    "tags are trimmed and empty entries ignored",
			tags:    []string{" sqli ", "", "  "},
			wantIDs: []string{"vt-dvwa"},
		},
		{
			name:    "filter and tags combine with AND",
			filter:  "goat",
			tags:    []string{"owasp"},
			wantIDs: []string{"vt-webgoat"},
		},
		{
			name:    "filter and tags with no common match returns empty",
			filter:  "dvwa",
			tags:    []string{"owasp"},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFiltered(templates, tt.filter, tt.tags)
			gotIDs := make([]string, 0, len(got))
			for _, tmpl := range got {
				gotIDs = append(gotIDs, tmpl.ID)
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}
