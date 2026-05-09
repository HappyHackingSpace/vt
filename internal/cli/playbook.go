package cli

import (
	"fmt"
	"strings"

	tmpl "github.com/happyhackingspace/vt/pkg/template"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newPlaybookCommand creates the playbook command with its subcommands.
func (c *CLI) newPlaybookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playbook",
		Short: "Manage and run playbooks (collections of templates)",
	}

	cmd.AddCommand(c.newPlaybookRunCommand())
	cmd.AddCommand(c.newPlaybookStopCommand())
	cmd.AddCommand(c.newPlaybookListCommand())
	return cmd
}

// newPlaybookRunCommand creates the playbook run subcommand.
func (c *CLI) newPlaybookRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run all templates defined in a playbook file",
		Run: func(cmd *cobra.Command, _ []string) {
			playbookID, err := cmd.Flags().GetString("id")
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			providerName, err := cmd.Flags().GetString("provider")
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			provider, ok := c.app.GetProvider(providerName)
			if !ok {
				log.Fatal().Msgf("provider %s not found", providerName)
			}

			pb, err := tmpl.GetPlaybookByID(c.app.Playbooks, playbookID)
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			log.Info().Msgf("running playbook '%s' [%s] (%d templates)", pb.Info.Name, pb.ID, len(pb.Templates))

			var failed []string
			for _, templateID := range pb.Templates {
				template, err := tmpl.GetByID(c.app.Templates, templateID)
				if err != nil {
					log.Error().Msgf("skipping '%s': %v", templateID, err)
					failed = append(failed, templateID)
					continue
				}

				log.Info().Msgf("starting template '%s'", templateID)
				if err := provider.Start(template); err != nil {
					log.Error().Msgf("failed to start '%s': %v", templateID, err)
					failed = append(failed, templateID)
					continue
				}

				if len(template.PostInstall) > 0 {
					log.Info().Msgf("post-install instructions for '%s':", templateID)
					for _, instruction := range template.PostInstall {
						fmt.Printf("  %s\n", instruction)
					}
				}
			}

			if len(failed) > 0 {
				log.Warn().Msgf("playbook finished with errors — failed templates: %s", strings.Join(failed, ", "))
			} else {
				log.Info().Msgf("playbook '%s' completed successfully", pb.Info.Name)
			}
		},
	}

	cmd.Flags().String("id", "", "Specify a playbook ID to run")
	cmd.Flags().StringP("provider", "p", "docker-compose",
		fmt.Sprintf("Specify the provider (%s)", strings.Join(c.providerNames(), ", ")))

	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Fatal().Msgf("%v", err)
	}

	return cmd
}

// newPlaybookStopCommand creates the playbook stop subcommand.
func (c *CLI) newPlaybookStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop all templates defined in a playbook",
		Run: func(cmd *cobra.Command, _ []string) {
			playbookID, err := cmd.Flags().GetString("id")
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			providerName, err := cmd.Flags().GetString("provider")
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			provider, ok := c.app.GetProvider(providerName)
			if !ok {
				log.Fatal().Msgf("provider %s not found", providerName)
			}

			pb, err := tmpl.GetPlaybookByID(c.app.Playbooks, playbookID)
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			log.Info().Msgf("stopping playbook '%s' [%s] (%d templates)", pb.Info.Name, pb.ID, len(pb.Templates))

			var failed []string
			for _, templateID := range pb.Templates {
				template, err := tmpl.GetByID(c.app.Templates, templateID)
				if err != nil {
					log.Error().Msgf("skipping '%s': %v", templateID, err)
					failed = append(failed, templateID)
					continue
				}

				log.Info().Msgf("stopping template '%s'", templateID)
				if err := provider.Stop(template); err != nil {
					log.Error().Msgf("failed to stop '%s': %v", templateID, err)
					failed = append(failed, templateID)
				}
			}

			if len(failed) > 0 {
				log.Warn().Msgf("playbook stop finished with errors — failed templates: %s", strings.Join(failed, ", "))
			} else {
				log.Info().Msgf("playbook '%s' stopped successfully", pb.Info.Name)
			}
		},
	}

	cmd.Flags().String("id", "", "Specify a playbook ID to stop")
	cmd.Flags().StringP("provider", "p", "docker-compose",
		fmt.Sprintf("Specify the provider (%s)", strings.Join(c.providerNames(), ", ")))

	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Fatal().Msgf("%v", err)
	}

	return cmd
}

// newPlaybookListCommand creates the playbook list subcommand.
func (c *CLI) newPlaybookListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available playbooks",
		Run: func(_ *cobra.Command, _ []string) {
			tmpl.ListPlaybooks(c.app.Playbooks)
		},
	}
}
