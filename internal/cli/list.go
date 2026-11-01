package cli

import (
	"strings"

	tmpl "github.com/happyhackingspace/vt/pkg/template"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newListCommand creates the list command.
func (c *CLI) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Run: func(cmd *cobra.Command, _ []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.Error().Err(err).Msg("failed to get filter flag")
				return
			}

			tagsValue, err := cmd.Flags().GetString("tags")
			if err != nil {
				log.Error().Err(err).Msg("failed to get tags flag")
				return
			}

			var tags []string
			if strings.TrimSpace(tagsValue) != "" {
				tags = strings.Split(tagsValue, ",")
			}

			tmpl.ListFilteredTemplates(c.app.Templates, filter, tags)
		},
	}

	cmd.Flags().StringP("filter", "f", "",
		"Filter templates by case-insensitive substring match on template ID and name")
	cmd.Flags().StringP("tags", "t", "",
		"Filter templates by comma-separated tags (exact case-insensitive match, a template matches if it has any of the tags)")

	return cmd
}
