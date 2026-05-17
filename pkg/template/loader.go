package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// TemplateRemoteRepository is the remote URL for the templates repository.
const TemplateRemoteRepository string = "https://github.com/HappyHackingSpace/vt-templates"

// maxScanDepth limits recursive directory scanning to prevent infinite loops from circular symlinks.
const maxScanDepth = 10

// LoadRepo loads all templates and playbooks from the given repository path in a
// single directory scan. Clones the repo first if it doesn't exist locally.
func LoadRepo(repoPath string) (map[string]Template, map[string]Playbook, error) {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Info().Msg("Fetching templates for the first time")
		if err := cloneTemplatesRepo(repoPath, false); err != nil {
			return nil, nil, fmt.Errorf("failed to clone templates repository: %w", err)
		}
	}

	templates := make(map[string]Template)
	playbooks := make(map[string]Playbook)

	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory %s: %w", repoPath, err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") || !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(repoPath, entry.Name())

		if entry.Name() == playbooksDirName {
			pbs, err := loadPlaybooksFromDir(entryPath)
			if err != nil {
				return nil, nil, err
			}
			for id, pb := range pbs {
				playbooks[id] = pb
			}
			continue
		}

		categoryTemplates, err := loadTemplatesFromCategory(entryPath, entry.Name())
		if err != nil {
			return nil, nil, err
		}
		for id, tmpl := range categoryTemplates {
			templates[id] = tmpl
		}
	}

	return templates, playbooks, nil
}

// SyncTemplates downloads or updates all templates from the remote repository.
func SyncTemplates(repoPath string) error {
	log.Info().Msgf("cloning %s", TemplateRemoteRepository)
	if err := cloneTemplatesRepo(repoPath, true); err != nil {
		return fmt.Errorf("failed to sync templates: %w", err)
	}
	return nil
}

// GetDockerComposePath finds and returns the docker-compose file path for a given template ID.
// Returns the absolute path to the compose file and its working directory.
func GetDockerComposePath(templateID, repoPath string) (composePath string, workingDir string, err error) {
	dirEntries, err := os.ReadDir(repoPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range dirEntries {
		if strings.HasPrefix(entry.Name(), ".") || !entry.IsDir() {
			continue
		}

		templateDir := filepath.Join(repoPath, entry.Name(), templateID)
		if !isTemplateDirectory(templateDir) {
			categoryPath := filepath.Join(repoPath, entry.Name())
			found, err := findTemplateInCategory(categoryPath, templateID)
			if err != nil {
				log.Debug().Err(err).Msgf("failed to find template %q in category %q", templateID, categoryPath)
				continue
			}
			if found == "" {
				continue
			}
			templateDir = found
		}

		tmpl, err := loadTemplate(templateDir)
		if err != nil {
			log.Debug().Err(err).Msgf("failed to load template %q from directory %q", templateID, templateDir)
			continue
		}

		if tmpl.ID != templateID {
			continue
		}

		providerConfig, exists := tmpl.Providers["docker-compose"]
		if !exists {
			return "", "", fmt.Errorf("template %q missing docker-compose provider configuration", templateID)
		}
		if providerConfig.Path == "" {
			return "", "", fmt.Errorf("template %q docker-compose.path is empty", templateID)
		}

		if filepath.IsAbs(providerConfig.Path) {
			return providerConfig.Path, filepath.Dir(providerConfig.Path), nil
		}

		composePath = filepath.Join(templateDir, providerConfig.Path)
		if _, statErr := os.Stat(composePath); statErr != nil {
			return "", "", fmt.Errorf("template %q has invalid docker-compose path %q: %w", templateID, composePath, statErr)
		}

		return composePath, filepath.Dir(composePath), nil
	}

	return "", "", fmt.Errorf("docker-compose file for template %q not found", templateID)
}

// loadTemplatesFromCategory loads all templates within a single category directory.
func loadTemplatesFromCategory(categoryPath, categoryName string) (map[string]Template, error) {
	templates := make(map[string]Template)

	err := filepath.WalkDir(categoryPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == categoryPath {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			log.Debug().Msgf("skipping symlink: %s", d.Name())
			return filepath.SkipDir
		}

		if !d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(categoryPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		if strings.Count(relPath, string(filepath.Separator))+1 > maxScanDepth {
			return fmt.Errorf("maximum directory depth (%d) exceeded at %s", maxScanDepth, path)
		}

		if isTemplateDirectory(path) {
			tmpl, err := loadTemplate(path)
			if err != nil {
				return fmt.Errorf("error loading template %s: %w", d.Name(), err)
			}
			if tmpl.ID != d.Name() {
				return fmt.Errorf("template id '%s' and directory name '%s' should match", tmpl.ID, d.Name())
			}
			if existing, exists := templates[tmpl.ID]; exists {
				return fmt.Errorf("duplicate template id '%s' found (already loaded: %s)", tmpl.ID, existing.Info.Name)
			}
			templates[tmpl.ID] = tmpl
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning category %s: %w", categoryName, err)
	}

	return templates, nil
}

// findTemplateInCategory recursively searches for a template directory within a category.
func findTemplateInCategory(categoryPath, templateID string) (string, error) {
	var foundPath string
	err := filepath.WalkDir(categoryPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != categoryPath {
			relPath, err := filepath.Rel(categoryPath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			if strings.Count(relPath, string(filepath.Separator))+1 > maxScanDepth {
				return fmt.Errorf("maximum directory depth (%d) exceeded at %s", maxScanDepth, path)
			}
		}

		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			return nil
		}

		if isTemplateDirectory(path) {
			if d.Name() == templateID {
				foundPath = path
				return filepath.SkipAll
			}
			return filepath.SkipDir
		}

		return nil
	})

	return foundPath, err
}

// isTemplateDirectory reports whether a directory contains an index.yaml file.
func isTemplateDirectory(dirPath string) bool {
	_, err := os.Stat(filepath.Join(dirPath, "index.yaml"))
	return err == nil
}
