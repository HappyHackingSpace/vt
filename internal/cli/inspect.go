package cli

import (
	"fmt"

	templ "github.com/happyhackingspace/vt/pkg/template"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (c *CLI) newInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "inspect operations",
		Run: func(cmd *cobra.Command, _ []string) {
			templateID, err := cmd.Flags().GetString("id")
			if err != nil {
				log.Fatal().Msgf("%v", err)
			}

			if template, err := templ.GetByID(c.app.Templates, templateID); err == nil {
				fmt.Println(template.String())
				return
			}

			if pb, err := templ.GetPlaybookByID(c.app.Playbooks, templateID); err == nil {
				fmt.Println(pb.String())
				return
			}

			log.Fatal().Msgf("'%s' not found as a template or playbook", templateID)
		},
	}
	cmd.Flags().String("id", "", "Specify a template ID for targeted vulnerable environment")
	err := cmd.MarkFlagRequired("id")
	if err != nil {
		panic(err)
	}
	return cmd
}
