package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTestPlaybook writes a playbook YAML file into pbDir named <id>.yaml.
func createTestPlaybook(t *testing.T, pbDir, id string, templates []string) {
	t.Helper()
	var sb strings.Builder
	fmt.Fprintf(&sb, "id: %s\ninfo:\n  name: Test Playbook %s\n  author: testauthor\ntemplates:\n", id, id)
	for _, tmpl := range templates {
		fmt.Fprintf(&sb, "  - %s\n", tmpl)
	}
	assert.NoError(t, os.MkdirAll(pbDir, 0750))
	assert.NoError(t, os.WriteFile(filepath.Join(pbDir, id+".yaml"), []byte(sb.String()), 0644))
}

func TestPlaybookValidate(t *testing.T) {
	tests := []struct {
		name    string
		pb      Playbook
		wantErr string
	}{
		{
			name:    "empty id",
			pb:      Playbook{Info: PlaybookInfo{Name: "x"}, Templates: []string{"a"}},
			wantErr: "id cannot be empty",
		},
		{
			name:    "invalid id format",
			pb:      Playbook{ID: "my-playbook", Info: PlaybookInfo{Name: "x"}, Templates: []string{"a"}},
			wantErr: "id must follow the format vt-pb-<number>",
		},
		{
			name:    "empty name",
			pb:      Playbook{ID: "vt-pb-001", Templates: []string{"a"}},
			wantErr: "info.name cannot be empty",
		},
		{
			name:    "empty templates",
			pb:      Playbook{ID: "vt-pb-001", Info: PlaybookInfo{Name: "x"}},
			wantErr: "templates list cannot be empty",
		},
		{
			name:    "blank template id",
			pb:      Playbook{ID: "vt-pb-001", Info: PlaybookInfo{Name: "x"}, Templates: []string{""}},
			wantErr: "template id cannot be empty",
		},
		{
			name:    "duplicate template id",
			pb:      Playbook{ID: "vt-pb-001", Info: PlaybookInfo{Name: "x"}, Templates: []string{"a", "a"}},
			wantErr: "duplicate template id 'a'",
		},
		{
			name: "valid",
			pb:   Playbook{ID: "vt-pb-001", Info: PlaybookInfo{Name: "x", Author: "a"}, Templates: []string{"a", "b"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pb.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestGetPlaybookByID(t *testing.T) {
	playbooks := map[string]Playbook{
		"vt-pb-001": {ID: "vt-pb-001", Info: PlaybookInfo{Name: "Lab One"}, Templates: []string{"tmpl-a"}},
	}

	pb, err := GetPlaybookByID(playbooks, "vt-pb-001")
	assert.NoError(t, err)
	assert.Equal(t, "vt-pb-001", pb.ID)

	pb, err = GetPlaybookByID(playbooks, "vt-pb-999")
	assert.Nil(t, pb)
	assert.EqualError(t, err, "playbook 'vt-pb-999' not found")
}

func TestLoadPlaybooksFromDir(t *testing.T) {
	tempDir := t.TempDir()

	createTestPlaybook(t, tempDir, "vt-pb-001", []string{"template-a", "template-b"})
	createTestPlaybook(t, tempDir, "vt-pb-002", []string{"template-c"})

	playbooks, err := loadPlaybooksFromDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, playbooks, 2)
	assert.Contains(t, playbooks, "vt-pb-001")
	assert.Contains(t, playbooks, "vt-pb-002")
	assert.Equal(t, []string{"template-a", "template-b"}, playbooks["vt-pb-001"].Templates)
}

func TestLoadPlaybooksFromDir_SkipsNonYAML(t *testing.T) {
	tempDir := t.TempDir()

	createTestPlaybook(t, tempDir, "vt-pb-001", []string{"template-a"})
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("ignore me"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "notes.md"), []byte("ignore me"), 0644))

	playbooks, err := loadPlaybooksFromDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, playbooks, 1)
	assert.Contains(t, playbooks, "vt-pb-001")
}

func TestLoadPlaybooksFromDir_IDFilenameMismatch(t *testing.T) {
	tempDir := t.TempDir()

	content := "id: vt-pb-999\ninfo:\n  name: Test\n  author: a\ntemplates:\n  - tmpl-a\n"
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "vt-pb-001.yaml"), []byte(content), 0644))

	_, err := loadPlaybooksFromDir(tempDir)
	assert.ErrorContains(t, err, "must match filename")
}

func TestLoadPlaybooksFromDir_DuplicateID(t *testing.T) {
	tempDir := t.TempDir()

	createTestPlaybook(t, tempDir, "vt-pb-001", []string{"template-a"})

	content := "id: vt-pb-001\ninfo:\n  name: Duplicate\n  author: a\ntemplates:\n  - template-b\n"
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "vt-pb-001.yml"), []byte(content), 0644))

	_, err := loadPlaybooksFromDir(tempDir)
	assert.ErrorContains(t, err, "duplicate playbook id")
}
