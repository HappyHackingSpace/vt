package template

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTestTemplate writes a minimal valid template directory under basePath.
func createTestTemplate(t *testing.T, basePath, templateID string) {
	t.Helper()
	content := fmt.Sprintf(`id: %s
info:
  name: Test Template %s
  author: testauthor
  description: Test description
  type: Lab
  targets:
    - test
  tags:
    - test
providers:
  docker-compose:
    path: "docker-compose.yaml"
`, templateID, templateID)

	dir := filepath.Join(basePath, templateID)
	assert.NoError(t, os.MkdirAll(dir, 0750))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "index.yaml"), []byte(content), 0644))
}

func TestIsTemplateDirectory(t *testing.T) {
	tempDir := t.TempDir()

	nonTemplateDir := filepath.Join(tempDir, "not-a-template")
	assert.NoError(t, os.MkdirAll(nonTemplateDir, 0750))
	assert.False(t, isTemplateDirectory(nonTemplateDir))

	templateDir := filepath.Join(tempDir, "is-a-template")
	assert.NoError(t, os.MkdirAll(templateDir, 0750))
	assert.NoError(t, os.WriteFile(filepath.Join(templateDir, "index.yaml"), []byte("test"), 0644))
	assert.True(t, isTemplateDirectory(templateDir))

	assert.False(t, isTemplateDirectory(filepath.Join(tempDir, "nonexistent")))
}

func TestLoadTemplatesFromCategoryNestedStructure(t *testing.T) {
	tempDir := t.TempDir()

	createTestTemplate(t, tempDir, "template-1")

	subdir := filepath.Join(tempDir, "subdir")
	assert.NoError(t, os.MkdirAll(subdir, 0750))
	createTestTemplate(t, subdir, "template-2")

	nested := filepath.Join(subdir, "nested")
	assert.NoError(t, os.MkdirAll(nested, 0750))
	createTestTemplate(t, nested, "template-3")

	templates, err := loadTemplatesFromCategory(tempDir, "test-category")
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Contains(t, templates, "template-1")
	assert.Contains(t, templates, "template-2")
	assert.Contains(t, templates, "template-3")
}

func TestLoadTemplatesFromCategorySkipsHiddenDirs(t *testing.T) {
	tempDir := t.TempDir()

	createTestTemplate(t, tempDir, "visible-template")

	hiddenDir := filepath.Join(tempDir, ".hidden")
	assert.NoError(t, os.MkdirAll(hiddenDir, 0750))
	createTestTemplate(t, hiddenDir, "hidden-template")

	templates, err := loadTemplatesFromCategory(tempDir, "test-category")
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Contains(t, templates, "visible-template")
	assert.NotContains(t, templates, "hidden-template")
}

func TestLoadTemplatesFromCategoryMaxDepth(t *testing.T) {
	tempDir := t.TempDir()

	currentPath := tempDir
	for i := 0; i <= maxScanDepth+1; i++ {
		currentPath = filepath.Join(currentPath, fmt.Sprintf("level%d", i))
		assert.NoError(t, os.MkdirAll(currentPath, 0750))
	}

	_, err := loadTemplatesFromCategory(tempDir, "test-category")
	assert.ErrorContains(t, err, "maximum directory depth")
}

func TestLoadTemplatesFromCategoryDuplicateID(t *testing.T) {
	tempDir := t.TempDir()

	createTestTemplate(t, tempDir, "duplicate-template")

	subdir := filepath.Join(tempDir, "subdir")
	assert.NoError(t, os.MkdirAll(subdir, 0750))
	createTestTemplate(t, subdir, "duplicate-template")

	_, err := loadTemplatesFromCategory(tempDir, "test-category")
	assert.ErrorContains(t, err, "duplicate template id")
}

func TestLoadRepo_TemplatesOnly(t *testing.T) {
	tempDir := t.TempDir()

	category1 := filepath.Join(tempDir, "category1")
	assert.NoError(t, os.MkdirAll(category1, 0750))
	createTestTemplate(t, category1, "template-a")

	category2 := filepath.Join(tempDir, "category2")
	assert.NoError(t, os.MkdirAll(category2, 0750))
	createTestTemplate(t, category2, "template-b")

	subgroup := filepath.Join(category2, "subgroup")
	assert.NoError(t, os.MkdirAll(subgroup, 0750))
	createTestTemplate(t, subgroup, "template-c")

	templates, playbooks, err := LoadRepo(tempDir)
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Contains(t, templates, "template-a")
	assert.Contains(t, templates, "template-b")
	assert.Contains(t, templates, "template-c")
	assert.Empty(t, playbooks)
}

func TestLoadRepo_WithPlaybooks(t *testing.T) {
	tempDir := t.TempDir()

	category := filepath.Join(tempDir, "category1")
	assert.NoError(t, os.MkdirAll(category, 0750))
	createTestTemplate(t, category, "template-a")
	createTestTemplate(t, category, "template-b")

	pbDir := filepath.Join(tempDir, playbooksDirName)
	createTestPlaybook(t, pbDir, "vt-pb-001", []string{"template-a", "template-b"})

	templates, playbooks, err := LoadRepo(tempDir)
	assert.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Len(t, playbooks, 1)
	assert.Contains(t, playbooks, "vt-pb-001")
}

func TestLoadRepo_SkipsHiddenDirs(t *testing.T) {
	tempDir := t.TempDir()

	visible := filepath.Join(tempDir, "visible")
	assert.NoError(t, os.MkdirAll(visible, 0750))
	createTestTemplate(t, visible, "visible-template")

	hidden := filepath.Join(tempDir, ".hidden")
	assert.NoError(t, os.MkdirAll(hidden, 0750))
	createTestTemplate(t, hidden, "hidden-template")

	templates, _, err := LoadRepo(tempDir)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Contains(t, templates, "visible-template")
	assert.NotContains(t, templates, "hidden-template")
}
