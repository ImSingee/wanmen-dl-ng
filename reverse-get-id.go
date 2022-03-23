package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var cmdReverseGetId = &cobra.Command{
	Use:   "get-id <name>",
	Short: "Use name get id",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("usage: get-id <name>")
		}

		id, ok := GetID(args[0])
		if !ok {
			return fmt.Errorf("not found")
		}

		fmt.Println(id)
		return nil
	},
}

func init() {
	app.AddCommand(cmdReverseGetId)
}
