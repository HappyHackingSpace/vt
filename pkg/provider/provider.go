// Package provider defines interfaces and types for managing vulnerable target environments.
package provider

import (
	"time"

	tmpl "github.com/happyhackingspace/vt/pkg/template"
)

// ListDeployment represents a running deployment discovered from the provider.
type ListDeployment struct {
	ProviderName string
	TemplateID   string
	Status       string
	CreatedAt    time.Time
}

// Provider defines the interface for managing vulnerable target environments.
type Provider interface {
	Name() string
	Start(template *tmpl.Template) error
	Stop(template *tmpl.Template) error
	Status(template *tmpl.Template) (string, error)
	List() ([]ListDeployment, error)
}
