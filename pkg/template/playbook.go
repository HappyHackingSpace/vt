package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var playbookIDRegex = regexp.MustCompile(`^vt-pb-\d+$`)

// PlaybookInfo contains metadata about a playbook.
type PlaybookInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
}

// Playbook groups multiple template IDs to run together.
type Playbook struct {
	ID        string       `yaml:"id"`
	Info      PlaybookInfo `yaml:"info"`
	Templates []string     `yaml:"templates"`
}

// playbooksDirName is the directory inside the template repo that holds playbook files.
const playbooksDirName = "playbooks"

// loadPlaybooksFromDir scans a playbooks directory and returns all playbooks indexed by their ID.
func loadPlaybooksFromDir(pbDir string) (map[string]Playbook, error) {
	playbooks := make(map[string]Playbook)

	entries, err := os.ReadDir(pbDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read playbooks directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		pb, err := loadPlaybookFromFile(filepath.Join(pbDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("error loading playbook '%s': %w", entry.Name(), err)
		}

		expectedID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if pb.ID != expectedID {
			return nil, fmt.Errorf("playbook id '%s' must match filename '%s'", pb.ID, expectedID)
		}

		if existing, exists := playbooks[pb.ID]; exists {
			return nil, fmt.Errorf("duplicate playbook id '%s' (already loaded: %s)", pb.ID, existing.Info.Name)
		}

		playbooks[pb.ID] = pb
	}

	return playbooks, nil
}

// GetPlaybookByID retrieves a playbook by its ID.
func GetPlaybookByID(playbooks map[string]Playbook, id string) (*Playbook, error) {
	pb, ok := playbooks[id]
	if !ok || pb.ID == "" {
		return nil, fmt.Errorf("playbook '%s' not found", id)
	}
	return &pb, nil
}

// ListPlaybooks prints all available playbooks as a table.
func ListPlaybooks(playbooks map[string]Playbook) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Author", "Templates"})

	for _, pb := range playbooks {
		t.AppendRow(table.Row{
			pb.ID,
			pb.Info.Name,
			pb.Info.Author,
			strings.Join(pb.Templates, "\n"),
		})
	}

	if t.Length() == 0 {
		fmt.Println("No playbooks found")
		return
	}

	t.SetCaption("there are %d playbooks", t.Length())
	t.SetIndexColumn(0)
	t.Render()
}

// String returns playbook fields as a table.
func (pb Playbook) String() string {
	tw := table.NewWriter()
	tw.AppendRow(table.Row{"ID", pb.ID})
	tw.AppendRow(table.Row{"Name", pb.Info.Name})
	tw.AppendRow(table.Row{"Author", pb.Info.Author})
	tw.AppendRow(table.Row{"Description", pb.Info.Description})
	tw.AppendRow(table.Row{"Templates", strings.Join(pb.Templates, "\n")})

	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateRows = true
	tw.Style().Options.SeparateColumns = true

	return tw.Render()
}

// Validate checks that the playbook has the required fields.
func (pb Playbook) Validate() error {
	if pb.ID == "" {
		return fmt.Errorf("playbook: id cannot be empty")
	}
	if !playbookIDRegex.MatchString(pb.ID) {
		return fmt.Errorf("playbook '%s': id must follow the format vt-pb-<number>", pb.ID)
	}
	if pb.Info.Name == "" {
		return fmt.Errorf("playbook '%s': info.name cannot be empty", pb.ID)
	}
	if len(pb.Templates) == 0 {
		return fmt.Errorf("playbook '%s': templates list cannot be empty", pb.ID)
	}
	seen := make(map[string]bool, len(pb.Templates))
	for _, id := range pb.Templates {
		if id == "" {
			return fmt.Errorf("playbook '%s': template id cannot be empty", pb.ID)
		}
		if seen[id] {
			return fmt.Errorf("playbook '%s': duplicate template id '%s'", pb.ID, id)
		}
		seen[id] = true
	}
	return nil
}
