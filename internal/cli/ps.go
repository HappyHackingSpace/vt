package cli

import (
	"os"
	"time"

	tmpl "github.com/happyhackingspace/vt/pkg/template"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newPsCommand creates the ps command.
func (c *CLI) newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List running deployments and their status",
		Run: func(cmd *cobra.Command, _ []string) {
			refresh, err := cmd.Flags().GetBool("refresh")
			if err != nil {
				log.Error().Msgf("%v", err)
				return
			}

			deployments, err := c.app.StateManager.ListDeployments()
			if err != nil {
				log.Error().Msgf("%v", err)
				return
			}

			t := table.NewWriter()
			t.SetStyle(table.StyleDefault)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Provider Name", "Template ID", "Status", "Created At"})

			count := 0
			for _, deployment := range deployments {
				provider, ok := c.app.GetProvider(deployment.ProviderName)
				if !ok {
					log.Error().Msgf("provider %q not found", deployment.ProviderName)
					continue
				}
				template, err := tmpl.GetByID(c.app.Templates, deployment.TemplateID)
				if err != nil {
					log.Error().Msgf("%v", err)
					continue
				}

				status := "unknown"
				if s, err := provider.Status(template); err != nil {
					log.Error().Msgf("%v", err)
				} else {
					status = s
				}

				if refresh && status != "running" {
					if err := c.app.StateManager.RemoveDeployment(deployment.ProviderName, deployment.TemplateID); err != nil {
						log.Error().Msgf("%v", err)
					} else {
						log.Info().Msgf("removed stale deployment %s on %s", deployment.TemplateID, deployment.ProviderName)
					}
					continue
				}

				t.AppendRow(table.Row{
					deployment.ProviderName,
					deployment.TemplateID,
					status,
					deployment.CreatedAt.Format(time.DateTime),
				})
				count++
			}

			if count == 0 {
				log.Info().Msg("there is no running environment")
				return
			}

			t.Render()
		},
	}

	cmd.Flags().BoolP("refresh", "r", false, "Remove stale deployments that are no longer running")

	return cmd
}
