package cli

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newPsCommand creates the ps command.
func (c *CLI) newPsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List running deployments and their status",
		Run: func(_ *cobra.Command, _ []string) {
			t := table.NewWriter()
			t.SetStyle(table.StyleDefault)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Provider Name", "Template ID", "Status", "Created At"})

			count := 0
			for _, provider := range c.app.Providers {
				deployments, err := provider.List()
				if err != nil {
					log.Error().Msgf("failed to list deployments from %s: %v", provider.Name(), err)
					continue
				}

				for _, deployment := range deployments {
					t.AppendRow(table.Row{
						deployment.ProviderName,
						deployment.TemplateID,
						deployment.Status,
						deployment.CreatedAt.Format(time.DateTime),
					})
					count++
				}
			}

			if count == 0 {
				log.Info().Msg("there is no running environment")
				return
			}

			t.Render()
		},
	}
}
