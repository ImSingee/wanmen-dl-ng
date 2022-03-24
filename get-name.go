package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var cmdGetName = &cobra.Command{
	Use:   "get-name <id>",
	Short: "Use id get name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("usage: get-name <id>")
		}

		name, ok := GetName(args[0])
		if !ok {
			return fmt.Errorf("not found")
		}

		fmt.Println(strings.TrimSpace(name))
		return nil
	},
}

func init() {
	app.AddCommand(cmdGetName)
}
