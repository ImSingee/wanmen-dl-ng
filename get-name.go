package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var cmdGetName = &cobra.Command{
	Use:   "get-name <id> ...",
	Short: "Use id get name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("usage: get-name <id>")
		}

		for _, id := range args {
			name, ok := GetName(id)
			if !ok {
				return fmt.Errorf("%s not found", id)
			}

			fmt.Println(strings.TrimSpace(name))
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdGetName)
}
