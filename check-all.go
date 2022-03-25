package main

import "github.com/spf13/cobra"

var cmdCheckAll = &cobra.Command{
	Use:     "check-all [<filename>]",
	Aliases: []string{"check-done", "ca"},
	Args:    cobraParseList,
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range list {
			checkDone(id)
		}
	},
}

func init() {
	app.AddCommand(cmdCheckAll)
}
