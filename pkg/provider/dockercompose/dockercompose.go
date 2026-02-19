// Package dockercompose provides Docker Compose provider implementation for managing vulnerable target environments.
package dockercompose

import (
	"fmt"
	"time"

	"github.com/happyhackingspace/vt/pkg/provider"
	tmpl "github.com/happyhackingspace/vt/pkg/template"
)

var _ provider.Provider = &DockerCompose{}

// DockerCompose implements the Provider interface using Docker Compose.
type DockerCompose struct{}

// NewDockerCompose creates a new DockerCompose provider.
func NewDockerCompose() *DockerCompose {
	return &DockerCompose{}
}

// Name returns the provider name.
func (d *DockerCompose) Name() string {
	return "docker-compose"
}

// isRunning checks if a template's containers are currently running via Docker API.
func (d *DockerCompose) isRunning(template *tmpl.Template) (bool, error) {
	dockerCli, err := createDockerCLI()
	if err != nil {
		return false, err
	}

	projectName := toProjectName(template.ID)
	containers, err := getProjectContainers(dockerCli, projectName)
	if err != nil {
		return false, err
	}

	if len(containers) == 0 {
		return false, nil
	}

	for _, c := range containers {
		if c.State == "running" {
			return true, nil
		}
	}

	return false, nil
}

// Start launches the vulnerable target environment using Docker Compose.
func (d *DockerCompose) Start(template *tmpl.Template) error {
	running, _ := d.isRunning(template) //nolint:errcheck
	if running {
		return fmt.Errorf("already running")
	}

	dockerCli, err := createDockerCLI()
	if err != nil {
		return err
	}

	project, err := loadComposeProject(*template)
	if err != nil {
		return err
	}

	return runComposeUp(dockerCli, project)
}

// Stop shuts down the vulnerable target environment using Docker Compose.
func (d *DockerCompose) Stop(template *tmpl.Template) error {
	running, err := d.isRunning(template)
	if err != nil {
		return err
	}

	if !running {
		return fmt.Errorf("deployment not running")
	}

	dockerCli, err := createDockerCLI()
	if err != nil {
		return err
	}

	project, err := loadComposeProject(*template)
	if err != nil {
		return err
	}

	return runComposeDown(dockerCli, project)
}

// Status returns status the vulnerable target environment using Docker Compose.
func (d *DockerCompose) Status(template *tmpl.Template) (string, error) {
	dockerCli, err := createDockerCLI()
	if err != nil {
		return "unknown", err
	}

	project, err := loadComposeProject(*template)
	if err != nil {
		return "unknown", err
	}

	running, err := runComposeStats(dockerCli, project)
	if err != nil {
		return "unknown", err
	}

	if !running {
		return "unknown", err
	}

	return "running", err
}

// List returns all running vt deployments discovered from Docker.
func (d *DockerCompose) List() ([]provider.ListDeployment, error) {
	dockerCli, err := createDockerCLI()
	if err != nil {
		return nil, err
	}

	stacks, err := listVTProjects(dockerCli)
	if err != nil {
		return nil, err
	}

	var deployments []provider.ListDeployment
	for _, stack := range stacks {
		// Get container details for template ID label and created time
		containers, err := getProjectContainers(dockerCli, stack.Name)
		if err != nil {
			continue
		}

		// Read template ID from container label (most reliable)
		// Falls back to project name conversion for older containers
		templateID := toTemplateID(stack.Name)
		for _, c := range containers {
			if id, ok := c.Labels["vt.template-id"]; ok {
				templateID = id
				break
			}
		}

		status := stack.Status
		var createdAt time.Time
		if len(containers) > 0 {
			// Use earliest container creation time
			createdAt = time.Unix(containers[0].Created, 0)
			for _, c := range containers[1:] {
				t := time.Unix(c.Created, 0)
				if t.Before(createdAt) {
					createdAt = t
				}
			}

			// Determine overall status from containers
			allRunning := true
			for _, c := range containers {
				if c.State != "running" {
					allRunning = false
					break
				}
			}
			if allRunning {
				status = "running"
			} else {
				// Skip deployments that are not fully running
				continue
			}
		}

		deployments = append(deployments, provider.ListDeployment{
			ProviderName: d.Name(),
			TemplateID:   templateID,
			Status:       status,
			CreatedAt:    createdAt,
		})
	}

	return deployments, nil
}
