package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var flagPrefix string
var flagShowId bool

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

			name = strings.TrimSpace(name)

			if flagShowId {
				fmt.Printf("%s%s %s\n", flagPrefix, id, name)
			} else {
				fmt.Printf("%s%s\n", flagPrefix, name)
			}
		}

		return nil
	},
}

func init() {
	app.AddCommand(cmdGetName)

	cmdGetName.Flags().BoolVarP(&flagShowId, "show-id", "i", false, "show id")
	cmdGetName.Flags().StringVarP(&flagPrefix, "prefix", "p", "", "prefix")
}
