package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var cmdCleanAll = &cobra.Command{
	Use:  "clean-all [<filename>]",
	Args: cobraParseList,
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range list {
			err := cleanById(id, flagDryRun)
			if err != nil {
				fmt.Printf("ERROR for %s: %v\n", id, err)
			}
		}
	},
}

func init() {
	app.AddCommand(cmdCleanAll)

	cmdCleanAll.Flags().BoolVar(&flagDryRun, "dry", false, "not really execute")
}
